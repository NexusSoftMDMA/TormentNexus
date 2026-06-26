"""Go code sanitization for LLM-generated tool implementations."""

import re

PROTECTED_FILES = {"registry.go", "parity.go", "factory.go"}


def sanitize_go_code(code: str) -> str:
    """Apply all sanitization steps to LLM-generated Go code."""

    # 0. Strip control characters from LLM output
    cleaned = []
    for ch in code:
        cp = ord(ch)
        if cp < 32 and ch not in ("\n", "\r", "\t"):
            continue
        cleaned.append(ch)
    code = "".join(cleaned)

    # 0b. Remove markdown code fences - KEEP content INSIDE fences
    lines = code.splitlines()
    has_fences = any(line.strip().startswith("```") for line in lines)
    if has_fences:
        cleaned_lines = []
        in_fence = False
        for line in lines:
            stripped = line.strip()
            if stripped.startswith("```"):
                in_fence = not in_fence
                continue
            if in_fence:
                cleaned_lines.append(line)
        code = "\n".join(cleaned_lines)

    # 0b2. Fix unterminated string literals
    # LLM sometimes generates strings with newlines in the middle.
    # Detect lines where quotes are opened but not closed.
    lines = code.splitlines()
    cleaned = []
    for line in lines:
        q_count = line.count('"') - line.count('\\"')
        if q_count % 2 != 0:
            # Odd number of quotes = unterminated string. Close it.
            line = line.rstrip() + '")'
        cleaned.append(line)
    code = "\n".join(cleaned)

    # 0c. Strip trailing non-Go text after the last function closing brace.
    # LLM sometimes adds attribution lines after the code.
    lines = code.splitlines()
    last_func_close = -1
    depth = 0
    for i, line in enumerate(lines):
        stripped = line.strip()
        depth += stripped.count("{") - stripped.count("}")
        if depth == 0 and stripped == "}" and i > 0:
            last_func_close = i
    if last_func_close >= 0 and last_func_close < len(lines) - 1:
        keep = True
        for j in range(last_func_close + 1, len(lines)):
            s = lines[j].strip()
            if s == "":
                continue
            if s.startswith("//") or s.startswith("/*"):
                continue
            if re.match(r"^(func |type |var |const )", s):
                continue
            keep = False
            break
        if not keep:
            code = "\n".join(lines[: last_func_close + 1])

    # 1. Replace package-level var/const apiBase/baseURL
    code = re.sub(
        r'^(var|const)\s+apiBase\s*=\s*"[^"]*"',
        r'\1 apiBase = "https://api.example.com"',
        code,
        flags=re.MULTILINE,
    )
    code = re.sub(
        r'^(var|const)\s+baseURL\s*=\s*"[^"]*"',
        r'\1 baseURL = "https://api.example.com"',
        code,
        flags=re.MULTILINE,
    )

    # 2. Replace httpClient with http.DefaultClient
    code = re.sub(r"\bhttpClient\b", "http.DefaultClient", code)
    code = re.sub(r":=\s*&http\.Client\{[^}]*\}", ":= http.DefaultClient", code)
    code = re.sub(r"\s*=\s*&http\.Client\{[^}]*\}", " = http.DefaultClient", code)
    code = re.sub(r"var\s+httpClient\s*=\s*&http\.Client\{[^}]*\}", "", code)
    code = code.replace(": =", ":=")

    # 3. Remove stray } after import closing )
    code = re.sub(r"\)\s*\n\s*\}\s*\n", ")\n\n", code)

    # 4. Combined brace fix: close if blocks after return, close functions before next func
    lines = code.splitlines()
    new_lines = []
    brace_depth = 0
    comment_buffer = []
    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped.startswith("//") or (stripped == "" and comment_buffer):
            comment_buffer.append((i, line))
            continue
        is_func = bool(re.match(r"^func\s+\w+\s*\(", stripped))
        if is_func and brace_depth > 0:
            while brace_depth > 0:
                new_lines.append("}")
                brace_depth -= 1
            new_lines.append("")
        for idx, buf_line in comment_buffer:
            new_lines.append(buf_line)
        comment_buffer = []
        brace_depth += stripped.count("{") - stripped.count("}")
        new_lines.append(line)
        if brace_depth > 0 and stripped.startswith("return"):
            next_nonblank = ""
            for j in range(i + 1, min(i + 3, len(lines))):
                ns = lines[j].strip()
                if ns:
                    next_nonblank = ns
                    break
            if next_nonblank and not next_nonblank.startswith("}"):
                new_lines.append("}")
                brace_depth -= 1
    for idx, buf_line in comment_buffer:
        new_lines.append(buf_line)
    while brace_depth > 0:
        new_lines.append("}")
        brace_depth -= 1
    code = "\n".join(new_lines)

    # 4c. Move } lines after func-describing comments to before them
    lines = code.splitlines()
    final = []
    i = 0
    while i < len(lines):
        if lines[i].strip().startswith("// Handle") or lines[i].strip().startswith(
            "// func "
        ):
            comment_start = i
            while comment_start > 0 and (
                lines[comment_start - 1].strip().startswith("//")
                or lines[comment_start - 1].strip() == ""
            ):
                comment_start -= 1
                if lines[comment_start].strip() == "" and comment_start < i - 1:
                    break
            closing_lines = []
            j = i + 1
            while j < len(lines) and lines[j].strip() == "}":
                closing_lines.append(lines[j])
                j += 1
            while j < len(lines) and lines[j].strip() == "":
                j += 1
            next_line = lines[j].strip() if j < len(lines) else ""
            if closing_lines and next_line.startswith("func "):
                for cl in closing_lines:
                    final.append(cl)
                final.append("")
                for k in range(comment_start, i + 1):
                    final.append(lines[k])
                i = j
                continue
        final.append(lines[i])
        i += 1
    code = "\n".join(final)

    # 5. Fix 2-value returns from getString/getInt/getBool
    for fn_name in ["getString", "getInt", "getBool"]:
        pattern = r"(\w+)\s*,\s*\w+\s*:=\s*" + fn_name + r"\("
        replacement = r"\1 := " + fn_name + r"("
        code = re.sub(pattern, replacement, code)

    # 5a. Fix single-value assignment from getString/getInt/getBool.
    # These functions return (value, bool). When used as 'x := getString(...)',
    # change to 'x, _ := getString(...)'.
    # But NOT when already multi-value: 'a, b := getString(...)' is fine.
    for fn_name in ["getString", "getInt", "getBool"]:
        lines = code.splitlines()
        fixed = []
        for line in lines:
            # Only fix if there is exactly ONE variable before :=
            # Pattern: word := getString(  (no comma before :=)
            import re as re_mod

            m = re_mod.match(r"^(\s*)(\w+)\s*:=\s*" + fn_name + r"\(", line)
            if m:
                indent = m.group(1)
                var = m.group(2)
                line = indent + var + ", _ :=" + fn_name + "(" + line[m.end() :]
            fixed.append(line)
        code = chr(10).join(fixed)

    # 5b. Remove error checks after getString/getInt/getBool calls.
    # These helpers return single values, not (value, error).
    lines = code.splitlines()
    cleaned = []
    i = 0
    while i < len(lines):
        cleaned.append(lines[i])
        stripped = lines[i].strip()
        is_getter_call = bool(re.match(r"\w+\s*:=\s*get(String|Int|Bool)\(", stripped))
        if is_getter_call:
            # Look ahead: if the next lines are "if e != nil { return err(...) }", remove them
            j = i + 1
            while j < len(lines) and lines[j].strip() == "":
                j += 1
            if j < len(lines) and re.match(
                r"if\s+e\s*(!=|==)\s*nil\s*\{", lines[j].strip()
            ):
                # Skip the if block
                k = j
                brace_count = 0
                while k < len(lines):
                    s = lines[k].strip()
                    brace_count += s.count("{") - s.count("}")
                    k += 1
                    if brace_count <= 0:
                        break
                # Remove lines j to k-1 from output
                i = k - 1  # will be incremented at end of loop
        i += 1
    code = "\n".join(cleaned)

    # 6. Fix ok/success variable shadowing
    code = re.sub(r",\s*success\s*:=", ", found :=", code)
    code = re.sub(r",\s*ok\s*:=", ", found :=", code)
    code = re.sub(r";\s+ok\s*\{", "; found {", code)
    code = re.sub(r";\s+success\s*\{", "; found {", code)
    code = re.sub(r"\bif\s+ok\s*\{", "if found {", code)
    code = re.sub(r"\bif\s+!ok\s*\{", "if !found {", code)
    code = re.sub(r"\bif\s+success\s*\{", "if found {", code)
    code = re.sub(r"\bif\s+!success\s*\{", "if !found {", code)
    code = re.sub(r"\bok\s*:=\s+", "found := ", code)
    code = re.sub(r"\bsuccess\s*:=\s+", "found := ", code)

    # 7. Fix err variable shadowing
    lines = code.splitlines()
    result_lines = []
    for line in lines:
        nl = line
        nl = re.sub(r",\s*err\s*:=\s*", ", e := ", nl)
        nl = re.sub(r",\s*err\s*=\s*", ", e = ", nl)
        nl = re.sub(r"\berr\s*:=\s", "e := ", nl)
        nl = re.sub(r"\bif\s+err\s*!=\s*nil", "if e != nil", nl)
        nl = re.sub(r"\bif\s+err\s*==\s*nil", "if e == nil", nl)
        nl = re.sub(r";\s*err\s*!=\s*nil", "; e != nil", nl)
        nl = re.sub(r";\s*err\s*==\s*nil", "; e == nil", nl)
        nl = re.sub(r"return\s+([^,\n]+),\s*err\b(?!\()", r"return \1, e", nl)
        nl = re.sub(r"^\s*err\s*=\s+", "e = ", nl)
        nl = re.sub(r"\berr\.", "e.", nl)
        nl = re.sub(r"\(err\)", "(e)", nl)
        nl = re.sub(r"\(err,", "(e,", nl)
        nl = re.sub(r",\s*err\)", ", e)", nl)
        nl = re.sub(r"\berr\b(?!\()", "e", nl)
        result_lines.append(nl)
    code = "\n".join(result_lines)

    # 8. Fix multi-value returns from err/success/ok functions
    for func_name in ["err", "success", "ok"]:
        changed = True
        while changed:
            changed = False
            search = "return " + func_name + "("
            idx = code.find(search)
            while idx >= 0:
                depth = 0
                end = idx + len(search)
                for j in range(end, len(code)):
                    if code[j] == "(":
                        depth += 1
                    elif code[j] == ")":
                        if depth == 0:
                            end = j + 1
                            break
                        depth -= 1
                after = code[end : end + 10].strip()
                if after.startswith(", nil") or after.startswith(",nil"):
                    nil_end = code.find("nil", end) + 3
                    code = code[:end] + code[nil_end:]
                    changed = True
                    break
                idx = code.find(search, end)

    # 8b. Fix empty err() calls
    code = code.replace("err()", 'err("error")')

    # 9. Remove golang.org/x imports
    code = re.sub(r'\s*"[^"]*golang\.org/x[^"]*"\n?', "\n", code)

    # 9b. Remove util import - it doesn't exist
    # Be specific: only match import lines, not strings containing "util"
    code = re.sub(
        r'^\s*"[^"]*github\.com/tormentnexus/go/internal/tools/util[^"]*"\s*$',
        "",
        code,
        flags=re.MULTILINE,
    )
    # Also match short form: import "util"
    code = re.sub(r'^\s*import\s+"util"\s*$', "", code, flags=re.MULTILINE)

    # 9c. Fix util.Success -> ok, util.Execute -> inline
    code = re.sub(r"\butil\.Success\s*\(", "ok(", code)
    code = re.sub(r"\butil\.Execute\s*\(", "ok(", code)

    # 10. Remove parity.go redeclarations
    for pf in [
        "getString",
        "getInt",
        "getBool",
        "ok",
        "err",
        "success",
        "ToolResponse",
    ]:
        pattern = r"func\s+" + pf + r"\s*\([^)]*\)[^{]*\}"
        code = re.sub(pattern, "", code, flags=re.DOTALL)
        code = re.sub(r"type\s+" + pf + r"\s+[^{]*\}", "", code)
        code = re.sub(r"type\s+" + pf + r"\s+string", "", code)

    # 11. Remove duplicate Handle declarations
    seen = set()
    lines = code.splitlines()
    new_lines = []
    skip = False
    for line in lines:
        stripped = line.strip()
        m = re.match(r"^func\s+(Handle\w+)\s*\(", stripped)
        if m:
            name = m.group(1)
            if name in seen:
                skip = True
                continue
            seen.add(name)
            skip = False
        elif re.match(r"^func\s+\w+\s*\(", stripped):
            skip = False
        if not skip:
            new_lines.append(line)
    code = "\n".join(new_lines)

    # 12. Ensure package first
    code = re.sub(r"^[\s\n]*package\s+tools", "package tools", code, count=1)

    # 13. Clean blank lines
    code = re.sub(r"\n{3,}", "\n\n", code)

    # 14. Remove blank imports
    code = re.sub(r"^\s*_\s*$", "", code, flags=re.MULTILINE)

    # 15. Remove non-Go lines between imports and first func
    # Skip this step - it was removing valid code inside functions
    # lines = code.splitlines()
    # cleaned = []
    # in_import = False
    # import_done = False
    # for line in lines:
    #     stripped = line.strip()
    #     if stripped.startswith("import"):
    #         in_import = True
    #     if in_import and stripped == ")":
    #         in_import = False
    #         import_done = True
    #         cleaned.append(line)
    #         continue
    #     if in_import:
    #         cleaned.append(line)
    #         continue
    #     if import_done and not re.match(
    #         r"^(func |type |var |const |// |/\*|$)", stripped
    #     ):
    #         continue
    #     if re.match(r"^(func |type )", stripped):
    #         import_done = False
    #     cleaned.append(line)
    # code = "\n".join(cleaned)

    # 16. Clean blank lines again
    code = re.sub(r"\n{3,}", "\n\n", code)

    return code

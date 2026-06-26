package tools, handlers are HandleXxx, use ok/e from parity.go, no external packages, only stdlib. Let's list 4-5 simple handlers that make sense for this course repo.

First, let's think of relevant tools:
1. get_course_overview: Returns the course description, what you'll learn, etc. That's from the README.
2. list_course_modules: Returns the syllabus modules, their links, summaries.
3. get_setup_instructions: Returns the getting started steps, like cloning, installing uv, setting up MCP server, API, etc.
4. check_kubrick_status: Wait no, maybe start_kubrick? No, wait the original has make start-kubrick, but maybe a tool that returns the service endpoints? Or get_service_endpoints: returns the URLs for MCP server, API, UI.
5. search_course_content: Maybe a simple search that looks for keywords in the course docs? Wait but keep it simple, no external packages, so use stdlib only.

Wait let's make sure each handler is simple, compiles, no external deps. Let's outline each handler:

First, get_course_overview: Takes no args? Or maybe optional? Wait args are map[string]interface{}, but if no args needed, just ignore. Returns the course overview text from the README.

Then list_course_modules: Returns the syllabus table as a formatted string.

Then get_setup_instructions: Returns the getting started steps from GETTING_STARTED.md.

Then get_service_endpoints: Returns the default ports and URLs for the three services.

Then search_course_docs: Takes a "query" string arg, searches through the course content (we can hardcode the relevant docs as a string, then do a simple substring match, case-insensitive) and returns matching sections.

Wait that's 5 handlers, which is within 2-6. Perfect.

Now, let's code each. First, package tools. Then, for each handler:

First, get_course_overview: HandleGetCourseOverview. Args can be empty, so just return ok with the overview text.

Then list_course_modules: HandleListCourseModules. Returns the syllabus.

Then get_setup_instructions: HandleGetSetupInstructions. Returns the getting started steps.

Then get_service_endpoints: HandleGetServiceEndpoints. Returns the three service URLs.

Then search_course_docs: HandleSearchCourseDocs. Takes "query" arg, getString(args, "query"), then search through the hardcoded docs, return matches.

Wait need to make sure all imports are stdlib: context, fmt, strings, maybe regexp? Wait no, for search, strings.Contains is enough, case-insensitive. So import strings, context, fmt.

Wait also, need to avoid shadowing e. So when doing error checks, use different names, but in these handlers, do we have any errors? Let's see: getString returns a string, no error? Wait wait, the user said getString returns single value. Oh right, the user said: "getString returns single value: val := getString(args, "key")". So no error from getString. So the only errors would be if, say, the query is empty? Wait no, even if query is empty, return all? Or return an error? Wait no, let's make it so if query is empty, return an error? Wait no, let's see: for HandleSearchCourseDocs, if query is empty, return err("query parameter is required")? Wait but e is the function from parity.go, right? So yes, return err("message").

Wait let's structure each handler:

First, HandleGetCourseOverview:
func HandleGetCourseOverview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    overview := `Kubrick Course: Learn to build AI Agents that can understand images, text, audio and videos.
This is a free, open-source course by The Neural Maze and Neural Bits, in collaboration with Pixeltable and Opik.
What you'll learn:
- Use Pixeltable for multimodal data processing and stateful agents
- Create complex MCP servers using FastMCP: expose resources, prompts, and tools
- Apply prompt versioning to your MCP server
- Learn how to implement custom MCP clients for your agents
- Implement an MCP Tool Agent from scratch, using Llama 4 Scout and Maverick as the LLMs
- Use Opik for MCP prompt versioning
- Learn how to implement custom tracing and monitoring with Opik`
    return ok(overview)
}

That's simple, no args needed, no errors.

Next, HandleListCourseModules:
func HandleListCourseModules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    modules := `Course Syllabus:
Module 0: Course introduction and overview. Outlining the basic components.
Module 1: Diving into the core components of the multimodal processing pipeline, covering video, images, text, and audio data. Code: kubrick-mcp
Module 2: Building the Kubrick MCP Server with FastMCP. Covering MCP protocol (transport, communication), capabilities, tools, resources, and prompts. Adding Opik Integration and MCP Inspector for efficient debugging and local testing. Code: kubrick-mcp
Module 3: Building the Kubrick MCP Agent from scratch. Implementing tool use, routing, and memory. Code: kubrick-api
Module 4: Integrating the Agent with the UI, adding observability, and productionizing the system. Code: kubrick-ui`
    return ok(modules)
}

Also simple, no args.

Next, HandleGetSetupInstructions:
func HandleGetSetupInstructions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    instructions := `Getting Started with Kubrick:
1. Clone the repository:
   git clone https://github.com/multi-modal-ai/multimodal-agents-course.git
   cd multimodal-agents-course
2. Install uv (Python package manager) from https://docs.astral.sh/uv/getting-started/installation/
3. Set up the MCP Server: Navigate to the kubrick-mcp directory and follow its README.md steps
4. Set up the MCP Agent / Agent API: Navigate to the kubrick-api directory and follow its README.md steps
5. Start Kubrick from the repo root with: make start-kubrick
   This starts 3 services:
   - MCP Server: http://localhost:9090/
   - Agent API: http://localhost:8080/
   - UI: http://localhost:3000/
Open http://localhost:3000/ to use Kubrick.`
    return ok(instructions)
}

Good, no args.

Next, HandleGetServiceEndpoints:
func HandleGetServiceEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    endpoints := `Kubrick Service Endpoints (default):
- MCP Server: http://localhost:9090/
- Agent API: http://localhost:8080/
- UI: http://localhost:3000/`
    return ok(endpoints)
}

Perfect.

Then HandleSearchCourseDocs: This one takes a "query" arg. So first, get the query string: query := getString(args, "query"). Then, if query is empty, return err("query parameter is required"). Then, hardcode the course content as a string, split into sections, then search each section for the query (case-insensitive), collect matches, return them.

Wait let's code that:
First, the course content can be a big string, or a slice of sections. Let's make a slice of sections:
var courseSections = []string{
    "Kubrick Course Overview: Learn to build AI Agents that can understand images, text, audio and videos. Free open-source course by The Neural Maze and Neural Bits, collaboration with Pixeltable and Opik.",
    "Module 1: Multimodal Data Pipelines. Covers video, image, text, audio processing using Pixeltable. Code in kubrick-mcp directory.",
    "Module 2: MCP Servers for Video Search. Build Kubrick MCP Server with FastMCP, covers MCP protocol, tools, resources, prompts, Opik integration, MCP Inspector. Code in kubrick-mcp.",
    "Module 3: Building Tool Use MCP Agents. Build Kubrick MCP Agent from scratch, implement custom MCP clients, use Llama 4 Scout/Maverick, Opik prompt versioning, tracing. Code in kubrick-api.",
    "Module 4: Productionizing the System. Integrate agent with UI, add observability, deploy with Docker Compose. Code in kubrick-ui.",
    "Getting Started: Clone repo, install uv, set up MCP server (kubrick-mcp), set up Agent API (kubrick-api), start with make start-kubrick. Services run on ports 9090 (MCP), 8080 (API), 3000 (UI).",
}

Then, for the query, convert to lower case, loop through each section, if strings.Contains(sectionLower, queryLower), add to matches. Then if no matches, return ok("No matching content found for query: " + query). Else return ok with the joined matches.

Wait also, need to import strings, right? Because we use strings.Contains, strings.ToLower.

Wait let's code that handler:
func HandleSearchCourseDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query parameter is required")
}

    queryLower := strings.ToLower(query)
    var matches []string
    for _, section := range courseSections {
        if strings.Contains(strings.ToLower(section), queryLower) {
            matches = append(matches, section)

    }
    if len(matches) == 0 {
        return ok(fmt.Sprintf("No matching course content found for query: %s", query))
}

    result := fmt.Sprintf("Found %d matching sections for '%s':\n\n", len(matches), query)
    for i, match := range matches {
        result += fmt.Sprintf("%d. %s\n\n", i+1, match)

    return ok(result)
}

Wait that's good. Now, need to import fmt and strings, right? So the imports would be:
import (
    "context"
    "fmt"
    "strings"
)
Wait wait, do we need any other imports? Let's check: no, all these use only context, fmt, strings. All stdlib, perfect.

Wait wait, what about the courseSections variable? It's a package-level var, right? So we can declare it outside the handlers, at the package level.

Wait let's put it all together. Let's check for any issues:

- No external packages: yes, only stdlib.
- No redeclaration of types: we don't declare ToolResponse, ok, e, getString, etc., they are from parity.go, correct.
- No variable named e: in all handlers, we don't declare any e variables, we use return err(...) which is the function, correct.
- All functions end with return: yes, each handler returns either ok or e.
- getString is used correctly: in HandleSearchCourseDocs := getString(args, "query"), correct.
- 5 handlers, which is within 2-6, perfect.
- Each handler is simple, compiles.

Wait wait, let's check the handler names: the manifest needs to have handler_func as HandleX, right? Let's list the handlers:

1. Tool name: get_course_overview, handler func: HandleGetCourseOverview, description: "Returns the full overview of the Kubrick multimodal agents course, including learning objectives and sponsors."
2. Tool name: list_course_modules, handler func: HandleListCourseModules, description: "Returns the full course syllabus with all modules, their summaries, and associated code directories."
3. Tool name: get_setup_instructions, handler func: HandleGetSetupInstructions, description: "Returns step-by-step instructions to set up and run the full Kubrick system locally."
4. Tool name: get_service_endpoints, handler func: HandleGetServiceEndpoints, description: "Returns the default local URLs and ports for all Kubrick services (MCP server, Agent API, UI)."
5. Tool name: search_course_docs, handler func: HandleSearchCourseDocs, description: "Searches the Kubrick course documentation for content matching the provided query string."

That's perfect. Now, let's write the full code.

Wait wait, the package is tools, right? So first line: package tools.

Then the imports:
import (
    "context"
    "fmt"
    "strings"
)

Then the package-level courseSections var:
var courseSections = []string{
    "Kubrick Course Overview: Learn to build AI Agents that can understand images, text, audio and videos. This is a free, open-source course by The Neural Maze and Neural Bits, in collaboration with Pixeltable and Opik.",
    "Module 1: Multimodal Data Pipelines. Dives into core components of multimodal processing, covering video, images, text, and audio data handling. Associated code: kubrick-mcp directory.",
    "Module 2: Building MCP Servers for Video Search. Build the Kubrick MCP Server with FastMCP, covering MCP protocol (transport, communication), capabilities, tools, resources, and prompts. Includes Opik Integration and MCP Inspector for debugging. Associated code: kubrick-mcp directory.",
    "Module 3: Building Tool Use MCP Agents. Build the Kubrick MCP Agent from scratch, implement custom MCP clients, use Llama 4 Scout and Maverick LLMs, Opik prompt versioning, and custom tracing/monitoring. Associated code: kubrick-api directory.",
    "Module 4: Productionizing the System. Integrate the agent with the UI, add full observability, and deploy the system using Docker Compose. Associated code: kubrick-ui directory.",
    "Getting Started Steps: 1. Clone the repository from https://github.com/multi-modal-ai/multimodal-agents-course.git. 2. Install uv package manager. 3. Set up MCP Server in kubrick-mcp directory. 4. Set up Agent API in kubrick-api directory. 5. Start all services with 'make start-kubrick' from the repo root.",
    "Default Service Endpoints: MCP Server runs on http://localhost:9090/, Agent API on http://localhost:8080/, UI on http://localhost:3000/.",
}

Then each handler function:

First, HandleGetCourseOverview:
}
}

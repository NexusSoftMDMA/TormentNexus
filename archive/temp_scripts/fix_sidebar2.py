import re

with open("packages/ui/src/components/Sidebar.tsx", "r") as f:
    data = f.read()

# Remove all bad imports
data = re.sub(r'import \{\s*Target,\s*Database,\s*Code2,\s*\} from \'next/link\';', "import Link from 'next/link';", data)
data = re.sub(r'import \{\s*Target,\s*Database,\s*Code2,\s*\} from \'next/navigation\';', "import { usePathname } from 'next/navigation';", data)
data = re.sub(r'import \{\s*Target,\s*Database,\s*Code2,\s*\} from \'\.\./lib/utils\';', "import { cn } from '../lib/utils';", data)

# Add them safely to lucide-react if they aren't there
if 'Target' not in data and 'lucide-react' in data:
    data = data.replace('HistoryIcon\n} from "lucide-react";', 'HistoryIcon,\n  Target,\n  Database,\n  Code2\n} from "lucide-react";')

with open("packages/ui/src/components/Sidebar.tsx", "w") as f:
    f.write(data)

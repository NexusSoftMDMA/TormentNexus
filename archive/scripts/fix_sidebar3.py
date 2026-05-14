with open("packages/ui/src/components/Sidebar.tsx", "r") as f:
    data = f.read()

# Fix the duplicate imports
data = data.replace("""import {
  Target,
  Database,
  Code2, usePathname } from 'next/navigation';""", "import { usePathname } from 'next/navigation';")

data = data.replace("""import {
  Target,
  Database,
  Code2,
  cn } from '../lib/utils';""", "import { cn } from '../lib/utils';")

with open("packages/ui/src/components/Sidebar.tsx", "w") as f:
    f.write(data)

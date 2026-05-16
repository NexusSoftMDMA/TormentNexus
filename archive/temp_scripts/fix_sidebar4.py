with open("packages/ui/src/components/Sidebar.tsx", "r") as f:
    data = f.read()

data = data.replace("""import {
  Target,
  Database,
  Code2, cn } from '../lib/utils';""", "import { cn } from '../lib/utils';")

data = data.replace("""import {
  Target,
  Database,
  Code2, Link } from 'next/link';""", "import Link from 'next/link';")

with open("packages/ui/src/components/Sidebar.tsx", "w") as f:
    f.write(data)

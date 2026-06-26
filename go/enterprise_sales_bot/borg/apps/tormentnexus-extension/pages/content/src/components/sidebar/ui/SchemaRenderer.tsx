import type React from 'react';
import { Typography } from '../ui';
import { cn } from '@src/lib/utils';

interface SchemaRendererProps {
  schema: any;
  className?: string;
}

const SchemaRenderer: React.FC<SchemaRendererProps> = ({ schema, className }) => {
  if (!schema || typeof schema !== 'object') {
    return <div className="text-xs text-slate-500 italic">No schema available</div>;
  }

  // Handle both standard JSON schema and OpenAI function format
  const properties = schema.properties || (schema.parameters && schema.parameters.properties) || {};
  const required = schema.required || (schema.parameters && schema.parameters.required) || [];

  if (Object.keys(properties).length === 0) {
    return <div className="text-xs text-slate-500 italic">No arguments required</div>;
  }

  return (
    <div className={cn('overflow-x-auto', className)}>
      <table className="w-full text-left text-xs border-collapse">
        <thead>
          <tr className="border-b border-slate-200 dark:border-slate-700">
            <th className="py-2 pl-1 pr-4 font-semibold text-slate-700 dark:text-slate-300">Name</th>
            <th className="py-2 px-4 font-semibold text-slate-700 dark:text-slate-300">Type</th>
            <th className="py-2 px-4 font-semibold text-slate-700 dark:text-slate-300">Description</th>
          </tr>
        </thead>
        <tbody>
          {Object.entries(properties).map(([name, prop]: [string, any]) => (
            <tr
              key={name}
              className="border-b border-slate-100 dark:border-slate-800 last:border-0 hover:bg-slate-50 dark:hover:bg-slate-800/50">
              <td className="py-2 pl-1 pr-4 align-top">
                <div className="flex items-center gap-1.5">
                  <code className="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-primary-600 dark:text-primary-400 font-mono text-[11px]">
                    {name}
                  </code>
                  {required.includes(name) && (
                    <span className="text-[10px] text-red-500 font-medium" title="Required">
                      *
                    </span>
                  )}
                </div>
              </td>
              <td className="py-2 px-4 align-top">
                <span className="text-slate-600 dark:text-slate-400 font-mono text-[10px]">{prop.type || 'any'}</span>
                {prop.enum && <div className="mt-1 text-[9px] text-slate-500">One of: {prop.enum.join(', ')}</div>}
              </td>
              <td className="py-2 px-4 align-top">
                <p className="text-slate-600 dark:text-slate-400 leading-snug">
                  {prop.description || <span className="italic opacity-50">No description</span>}
                </p>
                {prop.default !== undefined && (
                  <div className="mt-1 text-[9px] text-slate-500">
                    Default: <code className="bg-slate-50 dark:bg-slate-900 px-1 rounded">{String(prop.default)}</code>
                  </div>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default SchemaRenderer;

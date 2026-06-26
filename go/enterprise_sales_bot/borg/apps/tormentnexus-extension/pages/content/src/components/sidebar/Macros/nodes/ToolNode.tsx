import React, { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Icon } from '../../ui';

export const ToolNode = memo(({ data, isConnectable }: NodeProps) => {
    return (
        <div className="bg-white dark:bg-slate-800 border-2 border-blue-500 dark:border-blue-400 rounded-lg shadow-md min-w-[200px] overflow-hidden">
            <div className="bg-blue-50 dark:bg-blue-900/30 p-2 flex items-center justify-between border-b border-blue-200 dark:border-blue-800">
                <div className="flex items-center gap-2">
                    <Icon name="wrench" size="sm" className="text-blue-600 dark:text-blue-400" />
                    <span className="text-sm font-semibold text-blue-700 dark:text-blue-300">Tool</span>
                </div>
            </div>

            <div className="p-3">
                <div className="text-sm font-medium text-slate-700 dark:text-slate-200 mb-1">
                    {data?.toolName ? String(data.toolName) : 'Select a Tool...'}
                </div>
                {!!data?.args && (
                    <div className="text-[10px] font-mono text-slate-500 dark:text-slate-400 truncate max-w-[180px]">
                        {JSON.stringify(data.args)}
                    </div>
                )}
            </div>

            <Handle
                type="target"
                position={Position.Top}
                isConnectable={isConnectable}
                className="w-3 h-3 bg-blue-500"
            />
            <Handle
                type="source"
                position={Position.Bottom}
                isConnectable={isConnectable}
                className="w-3 h-3 bg-blue-500"
            />
        </div>
    );
});

ToolNode.displayName = 'ToolNode';

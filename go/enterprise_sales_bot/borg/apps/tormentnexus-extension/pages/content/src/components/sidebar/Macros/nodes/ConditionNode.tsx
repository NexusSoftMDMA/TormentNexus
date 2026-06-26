import React, { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Icon } from '../../ui';

export const ConditionNode = memo(({ data, isConnectable }: NodeProps) => {
    return (
        <div className="bg-white dark:bg-slate-800 border-2 border-amber-500 dark:border-amber-400 rounded-lg shadow-md min-w-[200px] overflow-hidden">
            <div className="bg-amber-50 dark:bg-amber-900/30 p-2 flex items-center justify-between border-b border-amber-200 dark:border-amber-800">
                <div className="flex items-center gap-2">
                    <Icon name="git-merge" size="sm" className="text-amber-600 dark:text-amber-400" />
                    <span className="text-sm font-semibold text-amber-700 dark:text-amber-300">Condition</span>
                </div>
            </div>

            <div className="p-3">
                <div className="text-xs font-mono text-slate-600 dark:text-slate-300 mb-2 p-1.5 bg-slate-50 dark:bg-slate-900 rounded border border-slate-100 dark:border-slate-700">
                    {data?.expression ? String(data.expression) : 'true'}
                </div>
            </div>

            <Handle
                type="target"
                position={Position.Top}
                isConnectable={isConnectable}
                className="w-3 h-3 bg-amber-500"
            />

            <Handle
                type="source"
                position={Position.Bottom}
                id="true"
                style={{ left: '30%', background: '#22c55e' }}
                isConnectable={isConnectable}
                className="w-3 h-3"
            />
            <span className="absolute bottom-[-20px] left-[20%] text-[10px] text-green-600 font-bold">TRUE</span>

            <Handle
                type="source"
                position={Position.Bottom}
                id="false"
                style={{ left: '70%', background: '#ef4444' }}
                isConnectable={isConnectable}
                className="w-3 h-3"
            />
            <span className="absolute bottom-[-20px] left-[60%] text-[10px] text-red-600 font-bold">FALSE</span>
        </div>
    );
});

ConditionNode.displayName = 'ConditionNode';

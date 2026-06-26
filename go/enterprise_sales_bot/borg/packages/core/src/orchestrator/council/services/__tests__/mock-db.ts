import { vi } from 'vitest';
import fs from 'fs';

let dbStore: Record<string, any[]> = {};

export function resetMockDb() {
  dbStore = {};
}

function getTable(name: string) {
  if (!dbStore[name]) dbStore[name] = [];
  return dbStore[name];
}

class MockDrizzleQueryBuilder {
  private currentTable: string = '';
  private currentOp: string = '';
  public customResolveData: any[] | null = null;
  public isCountQuery: boolean = false;
  private _limit: number | null = null;
  private _offset: number | null = null;
  private _orderByDirection: string | null = null;

  constructor() {}

  insert(table: any) {
    this.currentOp = 'insert';
    this.currentTable = table.name;
    return this;
  }

  values(val: any) {
    if (this.currentOp === 'insert') {
      const arr = Array.isArray(val) ? val : [val];
      getTable(this.currentTable).push(...arr);
      console.log(`Inserted into ${this.currentTable}, new len: ${getTable(this.currentTable).length}`);
      return Promise.resolve(arr);
    }
    return Promise.resolve(val);
  }

  select(columns?: any) {
    this.currentOp = 'select';
    this.isCountQuery = false;
    this._limit = null;
    this._offset = null;
    this._orderByDirection = null;
    if (columns && Object.keys(columns).includes('count')) {
        this.isCountQuery = true;
    }
    return this;
  }

  from(table: any) {
    this.currentTable = table.name;
    // We return a thenable that resolves to the table
    const self = this;
    const chain: any = {
      where: vi.fn((cond: any): any => {
        let extractedStrings: string[] = [];
        try {
            const extract = (obj: any) => {
                if (!obj) return;
                if (typeof obj === 'string') {
                    if (obj.trim() && obj.trim() !== '=' && obj.trim() !== '(' && obj.trim() !== ')' && obj.trim() !== 'and' && obj !== 'is null') {
                        extractedStrings.push(obj);
                    }
                } else if (Array.isArray(obj)) {
                    obj.forEach(extract);
                } else if (typeof obj === 'object') {
                    Object.values(obj).forEach(extract);
                }
            };
            extract(cond);
            // Filter the current table rows by checking if all extracted strings exist in the row
            const rawRows = getTable(self.currentTable);
            let filtered = rawRows;
            
            if (extractedStrings.length > 0) {
                filtered = rawRows.filter(row => {
                    const rowStr = JSON.stringify(row);
                    return extractedStrings.every(s => rowStr.includes(s));
                });
            }
                
            // HACK for debate-history specific filters
            if (self.currentTable === 'councilDebatesTable' && cond) {
                const condStr = JSON.stringify(cond);
                if (condStr.includes('"approved"') && !condStr.includes('session')) {
                     filtered = filtered.filter(row => row.outcome === 'approved');
                }
                if (condStr.includes('0.8')) {
                     filtered = filtered.filter(row => row.consensus >= 0.8);
                }
                if (condStr.includes('0.5')) {
                     filtered = filtered.filter(row => row.consensus <= 0.5);
                }
            }
            
            self.customResolveData = filtered;
        } catch(e) {}
        return chain;
      }),
      limit: vi.fn((n: number) => {
        self._limit = n;
        return chain;
      }),
      offset: vi.fn((n: number) => {
        self._offset = n;
        return chain;
      }),
      orderBy: vi.fn((direction: any) => {
        const dirStr = JSON.stringify(direction) || '';
        // Need to identify asc vs desc robustly
        if (dirStr.includes('"direction":"desc"') || (typeof direction === 'string' && direction.toLowerCase().includes('desc'))) {
            self._orderByDirection = 'desc';
        } else {
            self._orderByDirection = 'asc';
        }
        // Remember property to sort on
        (self as any)._orderByField = dirStr.includes('"consensus"') || typeof direction === 'string' && direction.includes('consensus') ? 'consensus' : 'timestamp';
        return chain;
      }),
      groupBy: vi.fn(() => chain),
      then: function(resolve: any) {
        let rowsToReturn = self.customResolveData || getTable(self.currentTable);
        self.customResolveData = null; // reset
        
        // Apply limit, offset, orderBy
        if (self._orderByDirection) {
            rowsToReturn = [...rowsToReturn].sort((a, b) => {
                const field = (self as any)._orderByField || 'timestamp';
                const valA = a[field] ?? a.timestamp;
                const valB = b[field] ?? b.timestamp;
                if (valA > valB) return self._orderByDirection === 'desc' ? -1 : 1;
                if (valA < valB) return self._orderByDirection === 'desc' ? 1 : -1;
                return 0;
            });
        }
        
        if (self._offset !== null) {
            rowsToReturn = rowsToReturn.slice(self._offset);
        }
        
        if (self._limit !== null) {
            rowsToReturn = rowsToReturn.slice(0, self._limit);
        }
        
        if (self.isCountQuery) {
             self.isCountQuery = false;
             return resolve([{ count: rowsToReturn.length }]);
        }
        return resolve(rowsToReturn);
      }
    };
    return chain;
  }

  delete(table: any) {
    this.currentOp = 'delete';
    this.currentTable = table.name;
    const self = this;
    let extractedStrings: string[] = [];
    
    const chain = {
        where: vi.fn((cond) => {
            const extract = (obj: any) => {
                if (!obj) return;
                if (typeof obj === 'string') {
                    if (obj.trim() && obj.trim() !== '=' && obj.trim() !== '(' && obj.trim() !== ')' && obj.trim() !== 'and' && obj !== 'is null') {
                        extractedStrings.push(obj);
                    }
                } else if (Array.isArray(obj)) {
                    obj.forEach(extract);
                } else if (typeof obj === 'object') {
                    Object.values(obj).forEach(extract);
                }
            };
            extract(cond);
            return Promise.resolve({ changes: self._executeDelete(extractedStrings) });
        }),
        then: function(resolve: any) {
            // Unconditional delete
            const rowCount = getTable(self.currentTable).length;
            dbStore[self.currentTable] = [];
            return resolve({ changes: rowCount });
        }
    };
    return chain;
  }

  private _executeDelete(extractedStrings: string[]): number {
      const rows = getTable(this.currentTable);
      const initialLength = rows.length;
      if (extractedStrings.length === 0) {
          dbStore[this.currentTable] = [];
          return initialLength;
      }
      
      const newRows = rows.filter(row => {
          const rowStr = JSON.stringify(row);
          // Keep rows that DO NOT match all extracted strings
          return !extractedStrings.every(s => rowStr.includes(s));
      });
      
      dbStore[this.currentTable] = newRows;
      return initialLength - newRows.length;
  }
  
  update(table: any) {
      this.currentOp = 'update';
      this.currentTable = table.name;
      const self = this;
      let setValues: any = {};
      let extractedStrings: string[] = [];

      const chain: any = {
          set: vi.fn((vals: any): any => {
              setValues = vals;
              return chain;
          }),
          where: vi.fn((cond: any) => {
              const extract = (obj: any) => {
                  if (!obj) return;
                  if (typeof obj === 'string') {
                      if (obj.trim() && obj.trim() !== '=' && obj.trim() !== '(' && obj.trim() !== ')' && obj.trim() !== 'and' && obj !== 'is null') {
                          extractedStrings.push(obj);
                      }
                  } else if (Array.isArray(obj)) {
                      obj.forEach(extract);
                  } else if (typeof obj === 'object') {
                      Object.values(obj).forEach(extract);
                  }
              };
              extract(cond);
              return chain;
          }),
          then: function(resolve: any) {
              const rows = getTable(self.currentTable);
              let changes = 0;
              for (let i = 0; i < rows.length; i++) {
                  const rowStr = JSON.stringify(rows[i]);
                  if (extractedStrings.length === 0 || extractedStrings.every(s => rowStr.includes(s))) {
                      rows[i] = { ...rows[i], ...setValues, _updatedByMock: true };
                      changes++;
                  }
              }
              return resolve({ changes });
          }
      };
      return chain;
  }
}

export const createDrizzleMock = () => {
  return new MockDrizzleQueryBuilder();
};

export const createSchemaMock = () => {
  return new Proxy({}, {
    get: (target, prop) => {
      return { name: prop as string };
    }
  });
};

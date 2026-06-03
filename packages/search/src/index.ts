// @tormentnexus/search — runtime-safe class stubs
// Provides class stubs that @tormentnexus/core uses as both types and values.
// Real implementations live in @tormentnexus/core runtime.

export interface SearchResult {
  title: string;
  url: string;
  snippet: string;
  content: string;
  file: string;
  line?: number;
  relevanceScore?: number;
}

export class SearchService {
  async search(_query: string, _root?: string, _opts?: any): Promise<SearchResult[]> { return []; }
  async execute(_input: any): Promise<any> { return { results: [] }; }
  async loadIndex(): Promise<void> {}
  getName(): string { return 'search'; }
}

package tools

import (  
    "context"  
)  

func SearchNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {  
    query, _ :=getString(args, "query")  
    _ = query  
    return ok("Searched for " + query)
}

func WriteNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {  
    content, _ :=getString(args, "content")  
    return ok("Written note: " + content)  
}// touch 1781132143

        import (
        	"context"
        	"encoding/json"
        	"fmt"
        	"io" // Added
        	"os"
        	"time"
        )
        if e == io.EOF {
            return []string{}, nil
        }
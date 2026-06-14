package tools

import (
    "context"
    "fmt"
)

func HandleSearchTickets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    from, _ :=getString(args, "from_station")
    to, _ :=getString(args, "to_station")
    date, _ :=getString(args, "date")
    if from == "" || to == "" || date == "" {
        return err("missing required parameters: from_station, to_station, date")
}

    return ok(fmt.Sprintf("Searching tickets from %s to %s on %s", from, to, date))
}

func HandleGetTrainInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    trainNo, _ :=getString(args, "train_no")
    if trainNo == "" {
        return err("missing required parameter: train_no")
}

    return ok(fmt.Sprintf("Retrieving info for train %s", trainNo))
}
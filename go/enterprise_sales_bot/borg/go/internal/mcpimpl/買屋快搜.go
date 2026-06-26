package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearchHouses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	district, _ :=getString(args, "district")
	minPrice, _ :=getInt(args, "min_price")
	maxPrice, _ :=getInt(args, "max_price")
	minRooms, _ :=getInt(args, "min_rooms")
	maxRooms, _ :=getInt(args, "max_rooms")
	msg := fmt.Sprintf("Searching houses in %s %s, price %d-%d, rooms %d-%d", city, district, minPrice, maxPrice, minRooms, maxRooms)
	return ok(msg)
}

func HandleGetHouseDetail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	houseID, _ :=getString(args, "house_id")
	if houseID == "" {
		return err("house_id is required")
}

	msg := fmt.Sprintf("Fetching detail for house %s", houseID)
	return ok(msg)
}
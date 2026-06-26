if fetchErr != nil {
    return err(fetchErr.Error())
}

defer resp.Body.Close()

responseBody, readErr := io.ReadAll(resp.Body)
if readErr != nil {
    return err(readErr.Error())
}
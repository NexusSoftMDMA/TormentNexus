        req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
        if e != nil {
            return err(e.Error())
}

        // ...
        resp, e := http.DefaultClient.Do(req)
        if e != nil {
            return err(e.Error())
}

        // ...
        if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
            return err(e.Error())
        }
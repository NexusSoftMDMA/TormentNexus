package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// OneSignal API configuration
func getOneSignalBaseURL() string {
	if u := os.Getenv("ONESIGNAL_API_URL"); u != "" {
		return u
	}
	return "https://onesignal.com/api/v1"
}

func getOneSignalAppID() string {
	return os.Getenv("ONESIGNAL_APP_ID")
}

func getOneSignalRestAPIKey() string {
	return os.Getenv("ONESIGNAL_REST_API_KEY")
}

// makeOneSignalRequest performs an HTTP request to the OneSignal API
func makeOneSignalRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	baseURL := getOneSignalBaseURL()
	fullURL := baseURL + endpoint

	req, reqErr := http.NewRequestWithContext(ctx, method, fullURL, body)
	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Authorization", "Basic "+getOneSignalRestAPIKey())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	return client.Do(req)
}

}

// HandleOneSignalCreateNotification creates a new push notification
func HandleOneSignalCreateNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID := getOneSignalAppID()
	if appID == "" {
		return err("ONESIGNAL_APP_ID environment variable not set")
}

	// Build notification payload
	payload := map[string]interface{}{
		"app_id": appID,
	}

	// Required: headings and contents or content
	if headings := getString(args, "headings"); headings != "" {
		payload["headings"] = map[string]string{"en": headings}
	}
	if contents := getString(args, "contents"); contents != "" {
		payload["contents"] = map[string]string{"en": contents}
	}

	// Optional fields
	if includedSegments := getString(args, "included_segments"); includedSegments != "" {
		payload["included_segments"] = []string{includedSegments}
	} else {
		payload["included_segments"] = []string{"All"}
	}

	if externalUserIDs := getString(args, "include_external_user_ids"); externalUserIDs != "" {
		payload["include_external_user_ids"] = []string{externalUserIDs}
	}

	if sendAfter := getString(args, "send_after"); sendAfter != "" {
		payload["send_after"] = sendAfter
	}

	if delayedOption := getString(args, "delayed_option"); delayedOption != "" {
		payload["delayed_option"] = delayedOption
	}

	if ttl := getInt(args, "ttl"); ttl > 0 {
		payload["ttl"] = ttl
	}

	if priority := getInt(args, "priority"); priority > 0 {
		payload["priority"] = priority
	}

	if data := getString(args, "data"); data != "" {
		var dataMap map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(data), &dataMap); jsonErr == nil {
			payload["data"] = dataMap
		}
	}

	if urlStr := getString(args, "url"); urlStr != "" {
		payload["url"] = urlStr
	}

	if webURL := getString(args, "web_url"); webURL != "" {
		payload["web_url"] = webURL
	}

	if iosBadgeType := getString(args, "ios_badgeType"); iosBadgeType != "" {
		payload["ios_badgeType"] = iosBadgeType
	}

	if iosBadgeCount := getInt(args, "ios_badgeCount"); iosBadgeCount > 0 {
		payload["ios_badgeCount"] = iosBadgeCount
	}

	if androidGroup := getString(args, "android_group"); androidGroup != "" {
		payload["android_group"] = androidGroup
	}

	if subtitle := getString(args, "subtitle"); subtitle != "" {
		payload["subtitle"] = map[string]string{"en": subtitle}
	}

	if largeIcon := getString(args, "large_icon"); largeIcon != "" {
		payload["large_icon"] = largeIcon
	}

	if bigPicture := getString(args, "big_picture"); bigPicture != "" {
		payload["big_picture"] = bigPicture
	}

	if buttons := getString(args, "buttons"); buttons != "" {
		var buttonsArray []map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(buttons), &buttonsArray); jsonErr == nil {
			payload["buttons"] = buttonsArray
		}
	}

	if isAndroid := getBool(args, "isAndroid"); isAndroid {
		payload["isAndroid"] = true
	}

	if isIos := getBool(args, "isIos"); isIos {
		payload["isIos"] = true
	}

	if isAnyWeb := getBool(args, "isAnyWeb"); isAnyWeb {
		payload["isAnyWeb"] = true
	}

	if isChromeWeb := getBool(args, "isChromeWeb"); isChromeWeb {
		payload["isChromeWeb"] = true
	}

	if isFirefox := getBool(args, "isFirefox"); isFirefox {
		payload["isFirefox"] = true
	}

	if isSafari := getBool(args, "isSafari"); isSafari {
		payload["isSafari"] = true
	}

	if isWP := getBool(args, "isWP"); isWP {
		payload["isWP"] = true
	}

	if isADM := getBool(args, "isADM"); isADM {
		payload["isADM"] = true
	}

	if isChrome := getBool(args, "isChrome"); isChrome {
		payload["isChrome"] = true
	}

	if isEdge := getBool(args, "isEdge"); isEdge {
		payload["isEdge"] = true
	}

	if isHuawei := getBool(args, "isHuawei"); isHuawei {
		payload["isHuawei"] = true
	}

	if huaweiMsg := getString(args, "huawei_msg"); huaweiMsg != "" {
		payload["huawei_msg"] = huaweiMsg
	}

	if existingKey := getString(args, "existing_key"); existingKey != "" {
		payload["existing_key"] = existingKey
	}

	if collapseID := getString(args, "collapse_id"); collapseID != "" {
		payload["collapse_id"] = collapseID
	}

	if threadID := getString(args, "thread_id"); threadID != "" {
		payload["thread_id"] = threadID
	}

	if summaryArg := getString(args, "summary_arg"); summaryArg != "" {
		payload["summary_arg"] = summaryArg
	}

	if summaryArgCount := getInt(args, "summary_arg_count"); summaryArgCount > 0 {
		payload["summary_arg_count"] = summaryArgCount
	}

	if emailSubject := getString(args, "email_subject"); emailSubject != "" {
		payload["email_subject"] = emailSubject
	}

	if emailBody := getString(args, "email_body"); emailBody != "" {
		payload["email_body"] = emailBody
	}

	if emailFromName := getString(args, "email_from_name"); emailFromName != "" {
		payload["email_from_name"] = emailFromName
	}

	if emailFromAddress := getString(args, "email_from_address"); emailFromAddress != "" {
		payload["email_from_address"] = emailFromAddress
	}

	if smsFrom := getString(args, "sms_from"); smsFrom != "" {
		payload["sms_from"] = smsFrom
	}

	if smsMediaUrls := getString(args, "sms_media_urls"); smsMediaUrls != "" {
		var urls []string
		if jsonErr := json.Unmarshal([]byte(smsMediaUrls), &urls); jsonErr == nil {
			payload["sms_media_urls"] = urls
		}
	}

	if filters := getString(args, "filters"); filters != "" {
		var filtersArray []map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(filters), &filtersArray); jsonErr == nil {
			payload["filters"] = filtersArray
		}
	}

	if includePlayerIDs := getString(args, "include_player_ids"); includePlayerIDs != "" {
		var ids []string
		if jsonErr := json.Unmarshal([]byte(includePlayerIDs), &ids); jsonErr == nil {
			payload["include_player_ids"] = ids
		}
	}

	if includeEmailTokens := getString(args, "include_email_tokens"); includeEmailTokens != "" {
		var tokens []string
		if jsonErr := json.Unmarshal([]byte(includeEmailTokens), &tokens); jsonErr == nil {
			payload["include_email_tokens"] = tokens
		}
	}

	if includePhoneNumbers := getString(args, "include_phone_numbers"); includePhoneNumbers != "" {
		var numbers []string
		if jsonErr := json.Unmarshal([]byte(includePhoneNumbers), &numbers); jsonErr == nil {
			payload["include_phone_numbers"] = numbers
		}
	}

	if includeIosTokens := getString(args, "include_ios_tokens"); includeIosTokens != "" {
		var tokens []string
		if jsonErr := json.Unmarshal([]byte(includeIosTokens), &tokens); jsonErr == nil {
			payload["include_ios_tokens"] = tokens
		}
	}

	if includeWpWnsUris := getString(args, "include_wp_uris"); includeWpWnsUris != "" {
		var uris []string
		if jsonErr := json.Unmarshal([]byte(includeWpWnsUris), &uris); jsonErr == nil {
			payload["include_wp_uris"] = uris
		}
	}

	if includeAmazonRegIds := getString(args, "include_amazon_reg_ids"); includeAmazonRegIds != "" {
		var ids []string
		if jsonErr := json.Unmarshal([]byte(includeAmazonRegIds), &ids); jsonErr == nil {
			payload["include_amazon_reg_ids"] = ids
		}
	}

	if includeChromeRegIds := getString(args, "include_chrome_reg_ids"); includeChromeRegIds != "" {
		var ids []string
		if jsonErr := json.Unmarshal([]byte(includeChromeRegIds), &ids); jsonErr == nil {
			payload["include_chrome_reg_ids"] = ids
		}
	}

	if includeAndroidRegIds := getString(args, "include_android_reg_ids"); includeAndroidRegIds != "" {
		var ids []string
		if jsonErr := json.Unmarshal([]byte(includeAndroidRegIds), &ids); jsonErr == nil {
			payload["include_android_reg_ids"] = ids
		}
	}

	if channelForExternalUserIds := getString(args, "channel_for_external_user_ids"); channelForExternalUserIds != "" {
		payload["channel_for_external_user_ids"] = channelForExternalUserIds
	}

	if apnsPushTypeOverride := getString(args, "apns_push_type_override"); apnsPushTypeOverride != "" {
		payload["apns_push_type_override"] = apnsPushTypeOverride
	}

	if name := getString(args, "name"); name != "" {
		payload["name"] = name
	}

	if appIds := getString(args, "app_ids"); appIds != "" {
		var ids []string
		if jsonErr := json.Unmarshal([]byte(appIds), &ids); jsonErr == nil {
			payload["app_ids"] = ids
		}
	}

	jsonBody, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err("Failed to marshal request body: " + marshalErr.Error())
}

	resp, apiErr := makeOneSignalRequest(ctx, "POST", "/notifications", stringReader(string(jsonBody)))
	if apiErr != nil {
		return err("OneSignal API request failed: " + apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("Failed to read response body: " + readErr.Error())
}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ok(string(bodyBytes))
}

	return err(fmt.Sprintf("OneSignal API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes)))
}

// stringReader creates an io.Reader from a string
func stringReader(s string) io.Reader {
	return &stringReaderImpl{s: s, i: 0}
}

type stringReaderImpl struct {
	s string
	i int
}

func (r *stringReaderImpl) Read(p []byte) (n int, e error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n = copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

// HandleOneSignalGetNotification retrieves information about a notification
func HandleOneSignalGetNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID := getOneSignalAppID()
	if appID == "" {
		return err("ONESIGNAL_APP_ID environment variable not set")
}

	notificationID, _ :=getString(args, "notification_id")
	if notificationID == "" {
		return err("notification_id is required")
}

	queryParams := url.Values{}
	queryParams.Set("app_id", appID)

	endpoint := fmt.Sprintf("/notifications/%s?%s", notificationID, queryParams.Encode())

	resp, apiErr := makeOneSignalRequest(ctx, "GET", endpoint, nil)
	if apiErr != nil {
		return err("OneSignal API request failed: " + apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("Failed to read response body: " + readErr.Error())
}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ok(string(bodyBytes))
}

	return err(fmt.Sprintf("OneSignal API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes)))
}

// HandleOneSignalCancelNotification cancels a scheduled notification
func HandleOneSignalCancelNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID := getOneSignalAppID()
	if appID == "" {
		return err("ONESIGNAL_APP_ID environment variable not set")
}

	notificationID, _ :=getString(args, "notification_id")
	if notificationID == "" {
		return err("notification_id is required")
}

	queryParams := url.Values{}
	queryParams.Set("app_id", appID)

	endpoint := fmt.Sprintf("/notifications/%s?%s", notificationID, queryParams.Encode())

	resp, apiErr := makeOneSignalRequest(ctx, "DELETE", endpoint, nil)
	if apiErr != nil {
		return err("OneSignal API request failed: " + apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("Failed to read response body: " + readErr.Error())
}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ok(string(bodyBytes))
}

	return err(fmt.Sprintf("OneSignal API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes)))
}

// HandleOneSignalViewNotifications lists all notifications for an app
func HandleOneSignalViewNotifications(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID := getOneSignalAppID()
	if appID == "" {
		return err("ONESIGNAL_APP_ID environment variable not set")
}

	queryParams := url.Values{}
	queryParams.Set("app_id", appID)

	if limit := getInt(args, "limit"); limit > 0 {
		queryParams.Set("limit", fmt.Sprintf("%d", limit))

	if offset := getInt(args, "offset"); offset > 0 {
		queryParams.Set("offset", fmt.Sprintf("%d", offset))

	if kind := getString(args, "kind"); kind != "" {
		queryParams.Set("kind", kind)

	endpoint := "/notifications?" + queryParams.Encode()

	resp, apiErr := makeOneSignalRequest(ctx, "GET", endpoint, nil)
	if apiErr != nil {
		return err("OneSignal API request failed: " + apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("Failed to read response body: " + readErr.Error())
}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ok(string(bodyBytes))
}

	return err(fmt.Sprintf("OneSignal API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes)))
}

}
}
}

// HandleOneSignalCreateSegment creates a new segment for an app
func HandleOneSignalCreateSegment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID := getOneSignalAppID()
	if appID == "" {
		return err("ONESIGNAL_APP_ID environment variable not set")
}

	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	payload := map[string]interface{}{
		"name": name,
	}

	if filters := getString(args, "filters"); filters != "" {
		var filtersArray []map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(filters), &filtersArray); jsonErr == nil {
			payload["filters"] = filtersArray
		}
	}

	jsonBody, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err("Failed to marshal request body: " + marshalErr.Error())
}

	endpoint := fmt.Sprintf("/apps/%s/segments", appID)

	resp, apiErr := makeOneSignalRequest(ctx, "POST", endpoint, stringReader(string(jsonBody)))
	if apiErr != nil {
		return err("OneSignal API request failed: " + apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("Failed to read response body: " + readErr.Error())
}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ok(string(bodyBytes))
}

	return err(fmt.Sprintf("OneSignal API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes)))
}

// HandleOneSignalDeleteSegment deletes a segment
func HandleOneSignalDeleteSegment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID := getOneSignalAppID()
	if appID == "" {
		return err("ONESIGNAL_APP_ID environment variable not set")
}

	segmentID, _ :=getString(args, "segment_id")
	if segmentID == "" {
		return err("segment_id is required")
}

	endpoint := fmt.Sprintf("/apps/%s/segments/%s", appID, segmentID)

	resp, apiErr := makeOneSignalRequest(ctx, "DELETE", endpoint, nil)
	if apiErr != nil {
		return err("OneSignal API request failed: " + apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
	}
}
package main

import (
    "io"
    "net/http"
    "log"
    "os"
    "strconv"
    "io/ioutil"
    "encoding/json"
    "errors"
)

const LISTEN_ADDRESS = ":9207"

var apiUrl string
var minerId string
var testMode string

type EwbfStatistics struct {
    Method string `json:"method"`
    Error string `json:"error"`
    StartTime int64 `json:"start_time"`
    CurrentServer string `json:"current_server"`
    AvailableServers int64 `json:"available_servers"`
    ServerStatus int64 `json:"server_status"`
    Result []struct {
        GpuId int64 `json:"gpuid"`
        CudaId int64 `json:"cudaid"`
        BusId string `json:"busid"`
        Name string `json:"name"`
        GpuStatus int64 `json:"gpu_status"`
        Solver int64 `json:"solver"`
        Temperature int64 `json:"temperature"`
        GpuPowerUsage int64 `json:"gpu_power_usage"`
        SpeedSps int64 `json:"speed_sps"`
        AcceptedShares int64 `json:"accepted_shares"`
        RejectedShares int64 `json:"rejected_shares"`
        Start_time int64 `json:"start_time"`
    } `json:"result"`
}

func integerToString(value int64) string {
    return strconv.FormatInt(value, 10)
}

func stringToFloat(value string) float64 {
    if value == "" {
        return 0
    }
    result, err := strconv.ParseFloat(value, 64)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func formatValue(key string, meta string, value string) string {
    result := key;
    if (meta != "") {
        result += "{" + meta + "}";
    }
    result += " "
    result += value
    result += "\n"
    return result
}

func queryData() (string, error) {
    var err error

    // Build URL
    url := apiUrl

    // Perform HTTP request
    resp, err := http.Get(url);
    if err != nil {
        return "", err;
    }

    // Parse response
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return "", errors.New("HTTP returned code " + integerToString(int64(resp.StatusCode)))
    }
    bodyBytes, err := ioutil.ReadAll(resp.Body)
    bodyString := string(bodyBytes)
    if err != nil {
        return "", err;
    }

    return bodyString, nil;
}

func getTestData() (string, error) {
    dir, err := os.Getwd()
    if err != nil {
        return "", err;
    }
    body, err := ioutil.ReadFile(dir + "/test.json")
    if err != nil {
        return "", err;
    }
    return string(body), nil
}

func metrics(w http.ResponseWriter, r *http.Request) {
    log.Print("Serving /metrics")

    var up int64 = 1
    var jsonString string
    var err error

    if (testMode == "1") {
        jsonString, err = getTestData()
    } else {
        jsonString, err = queryData()
    }
    if err != nil {
        log.Print(err)
        up = 0
    }

    // Parse JSON
    jsonData := EwbfStatistics{}
    json.Unmarshal([]byte(jsonString), &jsonData)

    if jsonData.Error != "" {
        log.Print("Response error: " + jsonData.Error)
        up = 0
    }

    // Sum stats of the GPUs
    var totalSpeedSps int64 = 0
    var totalAcceptedShares int64 = 0
    var totalRejectedShares int64 = 0
    for _, GPU := range jsonData.Result {
        totalSpeedSps += GPU.SpeedSps
        totalAcceptedShares += GPU.AcceptedShares
        totalRejectedShares += GPU.RejectedShares
    }

    // Output
    io.WriteString(w, formatValue("ewbf_up", "miner=\"" + minerId + "\"", integerToString(up)))
    io.WriteString(w, formatValue("ewbf_start_time", "miner=\"" + minerId + "\"", integerToString(jsonData.StartTime)))
    io.WriteString(w, formatValue("ewbf_speed_sps", "miner=\"" + minerId + "\"", integerToString(totalSpeedSps)))
    io.WriteString(w, formatValue("ewbf_accepted_shares", "miner=\"" + minerId + "\"", integerToString(totalAcceptedShares)))
    io.WriteString(w, formatValue("ewbf_rejected_shares", "miner=\"" + minerId + "\"", integerToString(totalRejectedShares)))
}

func index(w http.ResponseWriter, r *http.Request) {
    log.Print("Serving /index")
    html := `<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>EWBF Exporter</title>
    </head>
    <body>
        <h1>EWBF Exporter</h1>
        <p><a href="/metrics">Metrics</a></p>
    </body>
</html>`
    io.WriteString(w, html)
}

func main() {
    testMode = os.Getenv("TEST_MODE")
    if (testMode == "1") {
        log.Print("Test mode is enabled")
    }

    apiUrl = os.Getenv("API_URL")
    log.Print("API URL: " + apiUrl)

    minerId = os.Getenv("MINER_ID")
    log.Print("Miner ID: " + minerId)

    log.Print("EWBF exporter listening on " + LISTEN_ADDRESS)
    http.HandleFunc("/", index)
    http.HandleFunc("/metrics", metrics)
    http.ListenAndServe(LISTEN_ADDRESS, nil)
}

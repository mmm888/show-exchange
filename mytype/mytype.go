package mytype

var Exchange_data []string
var Exchange_map = map[string]int{"GBPNZD": 0, "CADJPY": 1, "GBPAUD": 2, "AUDJPY": 3, "AUDNZD": 4, "EURCAD": 5, "EURUSD": 6, "NZDJPY": 7, "USDCAD": 8, "EURGBP": 9, "GBPUSD": 10, "ZARJPY": 11, "EURCHF": 12, "CHFJPY": 13, "AUDUSD": 14, "USDCHF": 15, "EURJPY": 16, "GBPCHF": 17, "EURNZD": 18, "NZDUSD": 19, "USDJPY": 20, "EURAUD": 21, "AUDCHF": 22, "GBPJPY": 23}

type Exchange_db struct {
        time string
        open string
        bid  string
        ask  string
        high string
        low  string
}

type Gaitame struct {
        Quotes []struct {
                Code string `json:"currencyPairCode"`
                Open string `json:"open"`
                Bid  string `json:"bid"`
                Ask  string `json:"ask"`
                High string `json:"high"`
                Low  string `json:"low"`
        } `json:"quotes"`
        Time string
}

type TestTemplate struct {
        Main  string
        Sub   string
        Graph string
        Date  string
        Table string
}

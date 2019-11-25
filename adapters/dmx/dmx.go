package dmx

import (
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
)

const Seat = "dmx"
const RequestSize = 5
const endpoint = "https://dmx.districtm.io/b/v2"

type DmxAdapter struct {
}
type dmxSite struct {
	Domain  string      `json:"domain"`
	Page    string      `json:"page"`
	Content interface{} `json:content`
}

type dmxGeo struct {
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Type    int    `json:"type"`
	Country string `json:"country"`
}

type dmxDevice struct {
	Ua         string `json:"ua"`
	Devicetype int8   `json:"devicetype"`
	Ip         string `json:"ip"`
	Ipv6       string `json:"ipv6"`
	Geo        dmxGeo `json:"geo"`
	Language   string `json:"language"`
}

type dmxUser struct {
	Id      string `json:"id"`
	Buyerid string `json:"buyerid"`
}

type dmxImp struct {
	Id       string    `json:"id"`
	Banner   dmxBanner `json:"banner"`
	Tagid    string    `json:"tagid"`
	Bidfloor float32   `json:"bidfloor"`
	secure   int       `json:"secure"`
}

type dmxBanner struct {
	W     int8   `json:"w"`
	H     int8   `json:"h"`
	Battr []int8 `json:"battr"`
}

/**
openrtb.User should be use in that interface
*/
type dmxBannerRequest struct {
	Bcat   []string  `json:"bcat"`
	Badv   []string  `json:"badv"`
	Site   dmxSite   `json:"site"`
	Id     string    `json:"id"`
	Device dmxDevice `json:"device"`
	User   dmxUser   `json:"user"`
	Imp    []dmxImp  `json:"imp"`
}

func (dmx *DmxAdapter) MakeRequests(request *openrtb.BidRequest, reqData *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	// dmxRequests, _ := buildRequests(request)

	// fmt.Print(dmxRequests)

}

func (dmx *DmxAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {

}

func buildRequests(request *openrtb.BidRequest) (reqs dmxBannerRequest, errs []error) {
	var imps = make([]dmxImp, 0)
	reqs.Imp = imps

	return
}

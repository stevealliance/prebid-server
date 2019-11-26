package dmx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/prebid/prebid-server/errortypes"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const Seat = "dmx"
const RequestSize = 5

var endpoint = "https://dmx.districtm.io/b/v2"

type DmxAdapter struct {
}
type dmxSite struct {
	Domain    string       `json:"domain"`
	Page      string       `json:"page"`
	Content   interface{}  `json:content`
	Publisher dmxPublisher `json:"publisher"`
}

type dmxPublisher struct {
	ID string `json:"id"`
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
	IP         string `json:"ip"`
	Ipv6       string `json:"ipv6"`
	Geo        dmxGeo `json:"geo"`
	Language   string `json:"language"`
}

type dmxUser struct {
	ID      string `json:"id,omitempty"`
	BuyerID string `json:"buyerid,omitempty"`
}

type dmxImp struct {
	ID       string    `json:"id"`
	Banner   dmxBanner `json:"banner"`
	Tagid    string    `json:"tagid"`
	Bidfloor float32   `json:"bidfloor"`
	Secure   int8      `json:"secure"`
}

type dmxBanner struct {
	W      uint64           `json:"w"`
	H      uint64           `json:"h"`
	Battr  []uint64         `json:"battr,omitempty"`
	Format []openrtb.Format `json:"format,omitempty"`
}

type dmxFormat struct {
	W uint64 `json:"w"`
	H uint64 `json:"h"`
}

/**
openrtb.User should be use in that interface
*/
type dmxBannerRequest struct {
	Bcat   []string  `json:"bcat,omitempty"`
	Badv   []string  `json:"badv,omitempty"`
	Site   dmxSite   `json:"site"`
	ID     string    `json:"id"`
	Device dmxDevice `json:"device"`
	User   dmxUser   `json:"user"`
	Imp    []dmxImp  `json:"imp"`
}

type dmxResponse struct {
	ID      string   `json:"id"`
	ImpID   string   `json:"impid"`
	Price   float64  `json:"price"`
	NUrl    string   `json:"nurl,omitempty"`
	BUrl    string   `json:"burl,omitempty"`
	Adm     string   `json:"adm"`
	H       uint64   `json:"h"`
	W       uint64   `json:"w"`
	ADomain []string `json:"adomain"`
	Bundle  string   `json:"bundle,omitempty"`
	IUrl    string   `json:"iurl,omitempty"`
	Cid     string   `json:"cid"`
	Crid    string   `json:"crid"`
	Cat     []string `json:"cat,omitempty"`
}

func (dmx *DmxAdapter) MakeRequests(request *openrtb.BidRequest, reqData *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	fmt.Println("hello from steve 1")
	dmxRequests, _ := preprocess(request)
	errs := make([]error, 0)

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	reqArr := make([]*adapters.RequestData, 0)

	fmt.Println(dmxRequests)
	top := len(dmxRequests.Imp) > 0

	for top {
		resJSON, _ := json.Marshal(dmxRequests)
		reqArr = append(reqArr, &adapters.RequestData{
			Method:  "POST",
			Uri:     endpoint,
			Body:    resJSON,
			Headers: headers,
		})
	}
	return reqArr, errs
}

func (dmx *DmxAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code %d. Run with request.debug for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("Unexpected status code %d. Run with request.debug for more info", response.StatusCode)}
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	errs := make([]error, 0)
	resData := adapters.NewBidderResponseWithBidsCapacity(5)

	for _, bid := range bidResp.SeatBid {
		for i := 0; i < len(bid.Bid); i++ {
			b := bid.Bid[i]
			//var bidExt dmxResponse
			resData.Bids = append(resData.Bids, &adapters.TypedBid{
				Bid: &b,
			})
		}
	}

	return resData, errs
}

func preprocess(request *openrtb.BidRequest) (reqs dmxBannerRequest, errs []error) {

	var imps = make([]openrtb.Imp, 0)

	for i := 0; i < len(request.Imp); i++ {
		if request.Imp[i].Banner != nil && ((request.Imp[i].Banner.Format[0].H != 0 && request.Imp[i].Banner.Format[0].W != 0) ||
			(request.Imp[i].Banner.H != nil && request.Imp[i].Banner.W != nil)) {
			imps = append(imps, request.Imp[i])
		}
	}
	if len(imps) == 0 {
		errs = append(errs, errors.New("No valid impressions was found"))
		return
	}
	if len(imps) > 0 {
		request.Imp = imps
		reqs, errs = getBannerRequest(request)
	}

	return
}

func getBannerRequest(request *openrtb.BidRequest) (dmxBannerRequest, []error) {
	var dmx dmxBannerRequest
	var errs = make([]error, 0, len(request.Imp))
	var dmxExt openrtb_ext.ExtDmx
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(bidderExt.Bidder, &dmxExt); err != nil {
		errs = append(errs, err)
		return dmx, errs
	}
	dmx.Site = dmxSite{
		Domain: getDomain(request.Site.Page),
		Page:   request.Site.Page,
		Publisher: dmxPublisher{
			ID: dmxExt.MemberID,
		},
	}

	if request.User != nil && request.User.ID != "" {
		if dmx.User.ID == "" {
			dmx.User.ID = request.User.ID
		}
	}

	if request.User != nil && request.User.BuyerUID != "" {
		if dmx.User.BuyerID == "" {
			dmx.User.BuyerID = request.User.BuyerUID
		}
	}

	dmx.ID = request.ID
	Imp := make([]dmxImp, 0)
	for i := 0; i < len(request.Imp); i++ {

		// beachfrontExt, err := getBeachfrontExtension(request.Imp[i])

		// if err != nil {
		// 	errs = append(errs, err)
		// 	continue
		// }

		// appid, err := getAppId(beachfrontExt, openrtb_ext.BidTypeBanner)

		// if err != nil {
		// 	// Failed to get an appid, so this request is junk.
		// 	errs = append(errs, err)
		// 	continue
		// }
		var sizes dmxBanner

		sizes = dmxBanner{
			Format: request.Imp[i].Banner.Format,
			H:      request.Imp[i].Banner.Format[0].H,
			W:      request.Imp[i].Banner.Format[0].W,
		}

		Imp = append(Imp, dmxImp{
			ID:     request.Imp[i].ID,
			Secure: 1,
			Tagid:  dmxExt.TagID,
			Banner: sizes,
		})

		// if beachfrontExt.BidFloor <= minBidFloor {
		// 	slot.Bidfloor = 0
		// }

		// dmx.Imp = append(dmx.Imp, Imp)
	}
	dmx.Imp = Imp

	if len(dmx.Imp) == 0 {
		return dmx, errs
	}

	return dmx, errs
}

func getDomain(page string) string {
	protoURL := strings.Split(page, "//")
	var domainPage string

	if len(protoURL) > 1 {
		domainPage = protoURL[1]
	} else {
		domainPage = protoURL[0]
	}

	return strings.Split(domainPage, "/")[0]

}
func isSecure() int8 {
	return 1
}

func getIP(ip string) string {
	// This will only effect testing. The backend will return "" for localhost IPs,
	// and seems not to know what IPv6 is, so just setting it to one that is not likely to
	// be used.
	if ip == "" || ip == "::1" || ip == "127.0.0.1" {
		return "192.168.255.255"
	}
	return ip
}

func getBidType(externalRequest *adapters.RequestData) openrtb_ext.BidType {
	return openrtb_ext.BidTypeBanner
}

func NewDmxBidder() *DmxAdapter {
	return &DmxAdapter{}
}

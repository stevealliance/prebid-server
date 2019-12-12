package dmx

import (
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"net/http"
	"encoding/json"
	"fmt"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"sort"
)



type DmxAdapter struct {
	endpoint string
}

func NewDmxBidder(endpoint string) *DmxAdapter {
	return &DmxAdapter{endpoint: endpoint}
}

type dmxBidder struct {
	Bidder dmxExt `json:"bidder"`
}

type dmxExt struct {
	TagId string `json:"tagid,omitempty"`
	MemberId string `json:"memberid,omitempty"`
}

type dmxBanner struct {
	Banner *openrtb.Banner `json:"banner"`
}

type dmxSize struct {
	W uint64
	H uint64
	S uint64
}


type DmxSize []dmxSize

func (a DmxSize) Len() int {
	return len(a)
}

func (a DmxSize) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a DmxSize) Less(i, j int) bool {
	return a[i].S > a[j].S
}

func Remove(toBeRemove []openrtb.Format, a DmxSize) (dmx DmxSize) {
	for _, value := range toBeRemove {
		for _, dmxValue := range a {
			if dmxValue.H == value.H && dmxValue.W == value.W {
				fmt.Println("true")
				dmx = append(dmx, dmxValue)
			}
		}
	}
	return
}

var CheckTopSizes = []dmxSize{
	{300, 250, 100,},
	{728, 90, 95,},
	{320, 50, 90,},
	{160, 600, 88,},
	{300, 600, 85,},
	{300, 50, 80,},
	{970, 250, 75,},
	{970, 90, 70,},
}


func (adapter *DmxAdapter) MakeRequests(request *openrtb.BidRequest, req *adapters.ExtraRequestInfo) (reqsBidder []*adapters.RequestData, errs []error) {
	var dmxImp dmxBidder
	var imps []openrtb.Imp
	if err := json.Unmarshal(request.Imp[0].Ext, &dmxImp); err != nil {
		errs = append(errs, err)
	}


	for _, inst := range request.Imp {
		var banner openrtb.Banner
		var ins openrtb.Imp
		//for _, insbanner := range inst.Banner.Format {
		banner = openrtb.Banner{
			W: &inst.Banner.Format[0].W,
			H: &inst.Banner.Format[0].H,
			Format: inst.Banner.Format,
		}
		nSize := Remove(inst.Banner.Format, CheckTopSizes);
		sort.Sort(DmxSize(nSize))
		if inst.Banner.Format[0].W != 0 {
			banner.W = &nSize[0].W
		}
		if inst.Banner.Format[0].H != 0 {
			banner.H = &nSize[0].H
		}

		var intVal int8
		intVal = 1
		ins = openrtb.Imp{
			TagID: dmxImp.Bidder.TagId,
			ID: inst.ID,
			Banner: &banner,
			Ext: inst.Ext,
			Secure: &intVal,
		}
		imps = append(imps, ins)

	}

	request.Imp = imps

	request.Site.Publisher = &openrtb.Publisher{ ID: dmxImp.Bidder.MemberId,}
	oJson, _ := json.Marshal(request)
	headers := http.Header{}
	headers.Add("Content-Type", "Application/json;charset=utf-8")
	reqBidder := &adapters.RequestData{
		Method: "POST",
		Uri: "http://demo.arrepiblik.com/dmx2",//adapter.endpoint,
		Body: oJson,
		Headers: headers,
	}

	reqsBidder = append(reqsBidder, reqBidder)
	return
}

func (adapter *DmxAdapter) MakeBids(request *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errs []error

	if http.StatusNoContent == response.StatusCode {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("No content to be return"),
		}}
	}

	if http.StatusBadRequest == response.StatusCode {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Bad formated request"),
		}}
	}

	if http.StatusOK != response.StatusCode {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Something is really wrong"),
		}}
	}

	var bidResp openrtb.BidResponse

	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(5)

	for _, sb := range bidResp.SeatBid {
		for i := range sb.Bid {
			bidType, err := getMediaTypeForImp(sb.Bid[i].ImpID, request.Imp)
			if err != nil {
				errs = append(errs, err)
			} else {
				b := &adapters.TypedBid{
					Bid:     &sb.Bid[i],
					BidType: bidType,
				}
				bidResponse.Bids = append(bidResponse.Bids, b)
			}
		}
	}
	return bidResponse, errs

	return nil, errs
}

func getMediaTypeForImp(impID string, imps []openrtb.Imp) (openrtb_ext.BidType, error) {
	mediaType := openrtb_ext.BidTypeBanner
	for _, imp := range imps {
		if imp.ID == impID {
			if imp.Banner == nil && imp.Video != nil {
				mediaType = openrtb_ext.BidTypeVideo
			}
			return mediaType, nil
		}
	}

	// This shouldnt happen. Lets handle it just incase by returning an error.
	return "", &errortypes.BadInput{
		Message: fmt.Sprintf("Failed to find impression \"%s\" ", impID),
	}
}
package realestatecomau

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"github.com/PuerkitoBio/goquery"
	"errors"
	"strconv"
	"strings"
)

const (
	RealEstateComAuURL = "http://www.realestate.com.au"
	RealEstateComAuBuyURL = "http://www.realestate.com.au/buy/in-%s/list-1"
)

type Info struct {
	Address string
	PriceText string
	SaleType string
	PropertyType string
	Bedrooms int
	Bathrooms int
	CarSpaces int
	Link string
	FloorPlans []Image
	Photos []Image
}

type Image struct {
	DataType string
	ThumbURL string
	URL string
	Data []byte
}

func (info *Info) downloadImage(thumbURL, dataType string) (err error) {
	fields := strings.Split(thumbURL, "/")
	url := strings.Join(fields[: 3], "/") + "/5000x5000/" + strings.Join(fields[4 : len(fields)], "/")
	response, err := http.Get(url)

	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return
	}

	image := Image{
		URL: url,
		ThumbURL: thumbURL,
		DataType: dataType,
		Data: data,
	}

	if dataType == "floorplan" {
		info.FloorPlans = append(info.FloorPlans, image)
	} else {
		info.Photos = append(info.Photos, image)
	}

	return
}

func (info *Info) GetImages() (err error) {
	doc, err := goquery.NewDocument(info.Link)

	if err != nil {
		return
	}

	selection := doc.Find("#photoViewerCont > div.thumbs div.pages > div.page > div.thumb > img")

	selection.EachWithBreak(func(index int, selection *goquery.Selection) bool {
		var src string
		var dataType string

		for _, attr := range selection.Nodes[0].Attr {
			if attr.Key == "data-type" {
				dataType = attr.Val
			} else if attr.Key == "src" {
				src = attr.Val
			}
		}

		if src == "" || dataType == "" {
			err = errors.New("could not download photo")

			return false
		}

		if err = info.downloadImage(src, dataType); err != nil {
			return false
		}

		return true
	})

	return
}

func GetInfo(address string) (info *Info, err error) {
	url := fmt.Sprintf(RealEstateComAuBuyURL, url.QueryEscape(address))
	doc, err := goquery.NewDocument(url)

	if err != nil {
		return
	}

	info = &Info{
		Address: address,
	}

	selection := doc.Find("#resultsWrapper > p.noMatch").First()

	if len(selection.Nodes) == 1 {
		err = errors.New("No results found")

		return
	}

	selection = doc.Find("#searchResultsTbl > div.h1Wrapper > span").First()

	if selection.Text() == "No Exact Matches Found:" {
		err = errors.New("No exact results found")

		return
	}

	selection = doc.Find("div.resultBody")

	if len(selection.Nodes) != 1 {
		err = errors.New("Found more than one result")

		return
	}

	selection = doc.Find("div.resultBody div.propertyStats > p.price").First()

	if len(selection.Nodes) != 1 {
		err = errors.New("Could not parse price text")

		return
	}

	info.PriceText = selection.Text()

	selection = doc.Find("div.resultBody div.propertyStats > p.type").First()

	if len(selection.Nodes) != 1 {
		info.SaleType = "Unknown"
	} else {
		info.SaleType = selection.Text()
	}

	selection = doc.Find("div.resultBody div.listingInfo > span.propertyType").First()

	if len(selection.Nodes) != 1 {
		info.PropertyType = "Unknown"
	} else {
		info.PropertyType = selection.Text()
	}

	selection = doc.Find("div.resultBody div.listingInfo > ul.propertyFeatures > li")

	selection.Each(func(index int, selection *goquery.Selection) {
		img := selection.Find("img").First()
		num := selection.Find("span").First()

		for _, attr := range img.Nodes[0].Attr {
			if attr.Key == "alt" {
				var numValue int

				numValue, _ = strconv.Atoi(num.Text())

				if attr.Val == "Bedrooms" {
					info.Bedrooms = numValue
				} else if attr.Val == "Bathrooms" {
					info.Bathrooms = numValue
				} else if attr.Val == "Car Spaces" {
					info.CarSpaces = numValue
				}

				break
			}
		}
	})

	selection = doc.Find("div.resultBody div.vcard a").First()

	if len(selection.Nodes) != 1 {
		err = errors.New("Could not parse info link")

		return
	}

	for _, attr := range selection.Nodes[0].Attr {
		if attr.Key == "href" {
			info.Link = RealEstateComAuURL + attr.Val

			break
		}
	}

	return
}

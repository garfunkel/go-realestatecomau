// Package realestatecomau scrapes real estate information from realestate.com.au.
package realestatecomau

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// RealEstateComAuURL is the base URL for the website.
	RealEstateComAuURL = "http://www.realestate.com.au"

	// RealEstateComAuBuyURL is the URL used to view purchasable real estate.
	RealEstateComAuBuyURL = RealEstateComAuURL + "/buy/in-%s/list-1"
)

// Info is the structure containing scraped information.
type Info struct {
	Address      string
	PriceText    string
	SaleType     string
	PropertyType string
	Bedrooms     int
	Bathrooms    int
	CarSpaces    int
	Link         string
	FloorPlans   []Image
	Photos       []Image
	Inspections  []Inspection
}

// Image is the structure for images downloaded from the website.
type Image struct {
	DataType string
	ThumbURL string
	URL      string
	Data     []byte
}

// Inspection is a structure storing the inspection times for properties.
type Inspection struct {
	Date string
	Time string
}

// downloadImage is the method used to download images from the website given the thumbnail URL and type.
func (info *Info) downloadImage(thumbURL, dataType string) (err error) {
	fields := strings.Split(thumbURL, "/")
	url := strings.Join(fields[:3], "/") + "/5000x5000/" + strings.Join(fields[4:len(fields)], "/")
	response, err := http.Get(url)

	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return
	}

	image := Image{
		URL:      url,
		ThumbURL: thumbURL,
		DataType: dataType,
		Data:     data,
	}

	if dataType == "floorplan" {
		info.FloorPlans = append(info.FloorPlans, image)
	} else {
		info.Photos = append(info.Photos, image)
	}

	return
}

// GetImages downloads images for this property (see Photos and FloorPlans).
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

// GetInspections scrapes inspection details.
func (info *Info) GetInspections() (err error) {
	doc, err := goquery.NewDocument(info.Link)

	if err != nil {
		return
	}

	selection := doc.Find("#inspectionTimes > a.calendar-item")

	selection.EachWithBreak(func(index int, selection *goquery.Selection) bool {
		date := selection.Find("strong").First().Text()
		time := selection.Find("span.time").First().Text()

		if date == "" || time == "" {
			err = errors.New("could not parse inspection time info")

			return false
		}

		info.Inspections = append(info.Inspections, Inspection{Date: date, Time: time})

		return true
	})

	return
}

// GetInfo scrapes information for the address given. Images are not downloaded.
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

	selection = doc.Find("div.resultBody").First()

	if len(selection.Nodes) != 1 {
		err = errors.New("Unexpected number of results")

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

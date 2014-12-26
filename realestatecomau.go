package realestatecomau

import (
	"fmt"
	"net/url"
	"github.com/PuerkitoBio/goquery"
	"errors"
	"strconv"
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

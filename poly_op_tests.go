package main

import (
	"encoding/json"
	"fmt"
	//"github.com/akavel/polyclip-go"
	"github.com/ctessum/polyclip-go"
	"github.com/swill/go.clipper"
	"io/ioutil"
)

const (
	PRECISION float64 = 1000
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Points []Point

type TestCase struct {
	Subject Points `json:"subject"`
	Object  Points `json:"object"`
}

type TestCases []TestCase

func (ps Points) ToContour() polyclip.Contour {
	c := make(polyclip.Contour, 0)
	for i := range ps {
		c = append(c, polyclip.Point{ps[i].X, ps[i].Y})
	}
	return c
}

func ToPoints(cnt polyclip.Contour) Points {
	p := make(Points, 0)
	for i := range cnt {
		p = append(p, Point{cnt[i].X, cnt[i].Y})
	}
	return p
}

func (ps Points) ToPath() clipper.Path {
	p := make(clipper.Path, 0)
	for i := range ps {
		p = append(p, &clipper.IntPoint{clipper.CInt(ps[i].X * PRECISION), clipper.CInt(ps[i].Y * PRECISION)})
	}
	return p
}

func FromPath(cp clipper.Path) Points {
	p := make(Points, 0)
	for i := range cp {
		p = append(p, Point{float64(cp[i].X) / PRECISION, float64(cp[i].Y) / PRECISION})
	}
	return p
}

func main() {
	file, err := ioutil.ReadFile("./test_cases.json")
	if err != nil {
		panic(err)
	}

	var testcases TestCases
	err = json.Unmarshal(file, &testcases)
	if err != nil {
		panic(err)
	}

	success, failure := 0, 0
	for _, testcase := range testcases {
		poly_pts := polyclip.Polygon{testcase.Subject.ToContour()}
		poly_pts = poly_pts.Construct(polyclip.UNION, polyclip.Polygon{testcase.Object.ToContour()})
		if len(poly_pts) > 0 {
			success += 1
		} else {
			failure += 1
		}
	}

	fmt.Printf("Success Count: %d\n", success)
	fmt.Printf("Failure Count: %d\n", failure)

	success, failure = 0, 0
	for _, testcase := range testcases {
		c := clipper.NewClipper(clipper.IoNone)
		pft := clipper.PftEvenOdd

		c.AddPath(testcase.Subject.ToPath(), clipper.PtSubject, true)
		c.AddPath(testcase.Object.ToPath(), clipper.PtClip, true)

		solution, ok := c.Execute1(clipper.CtUnion, pft, pft)
		if !ok {
			failure += 1
		} else {
			if len(solution) == 1 {
				success += 1
			} else {
				failure += 1
			}
			fmt.Println("---")
			for _, s := range solution {
				for _, p := range s {
					fmt.Printf("(%d, %d), ", p.X, p.Y)
				}
				fmt.Printf("\n\n")
			}
		}
	}

	fmt.Printf("Success Count: %d\n", success)
	fmt.Printf("Failure Count: %d\n", failure)
}

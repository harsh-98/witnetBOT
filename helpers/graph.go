package helpers

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/golang/freetype"
	"github.com/harsh-98/witnetBOT/log"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

func getColors() []color.RGBA {
	clRed := color.RGBA{255, 0, 0, 0xff}
	clGreen := color.RGBA{0, 127, 0, 0xff}
	clBlue := color.RGBA{0, 0, 255, 0xff}
	clYellow := color.RGBA{255, 255, 0, 0xff}
	clAqua := color.RGBA{0, 255, 255, 0xff}
	// clTeal := color.RGBA{0, 127, 127, 0xff}
	// clSilver := color.RGBA{127, 127, 127, 0xff}
	clFuchsia := color.RGBA{255, 0, 255, 0xff}
	clOlive := color.RGBA{127, 127, 0, 0xff}
	clPurple := color.RGBA{127, 0, 127, 0xff}
	clWhite := color.RGBA{255, 255, 255, 0xff}
	var colors = []color.RGBA{clFuchsia, clOlive, clGreen, clBlue, clPurple, clRed, clWhite, clYellow, clAqua}
	return colors
}

type point struct {
	X time.Time
	Y float64
}

func loadFont(gc draw2d.GraphicContext) {
	// get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Logger.Fatal(err)
	}

	// parse the font file
	// https://github.com/llgcode/draw2d/issues/127#issuecomment-267845074
	// https://play.golang.org/p/MuxKhec9G9H
	s := fmt.Sprintf("%s/font/Tahomasr.tff", dir)
	fontBytes, err := ioutil.ReadFile(s)
	if err != nil {
		log.Logger.Error("Can not load font")

	}
	regularFont, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Logger.Error("Can not parse font")
	}

	// add font for rendering the label
	fondata := draw2d.FontData{Name: "Tahoma", Family: draw2d.FontFamilyMono, Style: draw2d.FontStyleNormal}
	draw2d.RegisterFont(
		draw2d.FontData{Name: "Tahoma", Family: draw2d.FontFamilyMono, Style: draw2d.FontStyleNormal},
		regularFont,
	)
	gc.SetFontData(fondata)
}

func addLabel(gc draw2d.GraphicContext, clr color.RGBA, text string, x, y float64) {
	gc.SetFontSize(14)
	gc.SetFillColor(clr)
	gc.FillStringAt(text, x, y)
}
func drawLine(gc draw2d.GraphicContext, x0, y0, x1, y1 float64) {
	// Draw a line
	// fmt.Println(x0, y0, x1, y1)
	gc.MoveTo(x0, y0)
	gc.LineTo(x1, y1)
	gc.Stroke()
}
func GenerateGraph(tgID int64) error {
	title := "Rating"

	// Total Time
	var graphXAxisDurationHr uint64 = uint64(Config.GetInt("graphXAxisDurationHr"))

	// min and max x coordinate
	maxX := time.Now().Unix()
	minX := maxX - int64(graphXAxisDurationHr*60*60)
	// get Colors
	colors := getColors()

	// nodes and its length
	nodeIDToName := global.Users[tgID].Nodes
	nodeIDs := []string{}
	for k, _ := range nodeIDToName {
		nodeIDs = append(nodeIDs, k)
	}
	nLen := len(nodeIDs)

	// reutrn in user hasn't added any users
	if nLen == 0 {
		msg := tgbotapi.NewMessage(tgID, "You haven't added nodes")
		TgBot.Send(msg)
		return nil
	}
	exitReturn := true
	for _, v := range nodeIDs {
		if global.NodeRepMap[v] != nil {
			exitReturn = false
		}
	}
	if exitReturn {
		msg := tgbotapi.NewMessage(tgID, "Your nodes are not in reputation list. I will notify if added 😄.")
		TgBot.Send(msg)
		return nil
	}

	// ?, ?, ?, ? and entries which will be replacing the question marks
	var qMarkArr []string
	var entries []interface{}
	for _, nodeID := range nodeIDs {
		qMarkArr = append(qMarkArr, "?")
		entries = append(entries, nodeID)
	}
	placeHolder := strings.Join(qMarkArr, ", ")
	query := fmt.Sprintf("select NodeID, Reputation, CreateAt from reputation where NodeID in (%s) and CreateAt > DATE_SUB(NOW(), INTERVAL %v HOUR);", placeHolder, graphXAxisDurationHr)
	log.Logger.Debugf("Graph query: %s", query)
	rows, err := sqldb.Query(query, entries...)
	if err != nil {
		log.Logger.Errorf("Error fetching rating from DB: %s\n\r", err)
		return err
	}
	var (
		nodeID       string
		createdAtStr string
		reputation   float64
	)
	graphData := make(map[string][]point)
	var maxY, minY int64
	minY = 1000000000

	// iter over row
	for rows.Next() {
		err := rows.Scan(&nodeID, &reputation, &createdAtStr)
		if err != nil {
			log.Logger.Errorf("Error reading row: %s\n", err)
			return nil
		}
		// RFC3339
		// https://yourbasic.org/golang/format-parse-string-time-date-example/
		createdAt, err := time.Parse(TIMEFORMAT, createdAtStr)
		// fmt.Println(createdAtStr)
		// fmt.Println(createdAt)
		if err != nil {
			log.Logger.Errorf("Error in parsing time: %s\n", err)
			return nil
		}
		graphData[nodeID] = append(graphData[nodeID], point{
			X: createdAt,
			Y: reputation,
		})
		y := int64(reputation)
		if y > maxY {
			maxY = y
		}
		if y < minY {
			minY = y
		}
	}
	rows.Close()
	if len(graphData) == 0 {
		msg := tgbotapi.NewMessage(tgID, fmt.Sprintf("Your nodes haven't been rated in last %v", graphXAxisDurationHr))
		TgBot.Send(msg)
		return nil
	}
	log.Logger.Debugf("User id %v with %v nodes with min  %v  and max %v reputation", tgID, nLen, minY, maxY)
	// set Y min , max coordinate
	diff := int64(float64(maxY-minY) * 0.05)
	maxY = maxY + diff
	minY = minY - diff

	// define the image boundaries
	var width float64 = 1920
	var height float64 = 1080
	upLeft := image.Point{0, 0}
	lowRight := image.Point{int(width), int(height)}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// define margin
	var leftMargin float64 = 100
	var rightMargin float64 = 100
	var topMargin float64 = 50
	var bottomMargin float64 = 60

	gc := draw2dimg.NewGraphicContext(img)
	log.Logger.Tracef("%+v\n", graphData)

	draw2dkit.Rectangle(gc, 0, 0, width, height)
	gc.SetFillColor(image.Black)
	gc.FillStroke()

	// Add gradient background
	for i := 0; i < int(height-topMargin-bottomMargin)/20; i++ {
		j := uint8(i / 2)
		draw2dkit.Rectangle(gc, leftMargin, topMargin+float64(i*20), width-rightMargin, topMargin+float64(i*20+22))
		gc.SetFillColor(color.RGBA{j + 5, j + 5, j + 5, 0xff})
		gc.FillStroke()
	}
	// load font
	loadFont(gc)

	// Set the gap between x and y labels
	xDisplayCount := 75
	yDisplayCount := 40

	// add title
	addLabel(gc, color.RGBA{0xff, 0xff, 0xff, 0xff}, title, (width-float64(20*len(title)))/2, 50)
	// add x axis label
	innerWidth := int(width-leftMargin-rightMargin) / xDisplayCount
	for i := 0; i <= innerWidth; i++ {
		t := int(maxX-minX)*i/innerWidth + int(minX)
		t /= 60
		t %= 60 * 24
		text := fmt.Sprintf("%02v:%02v", t/60, t%60)
		addLabel(gc, colors[7], text, float64(i*xDisplayCount-20)+leftMargin, height+float64(16-bottomMargin))
	}

	// add y axis label
	innerHeight := int(height-topMargin-bottomMargin) / yDisplayCount
	for i := 0; i <= innerHeight; i++ {
		r := int64(int(maxY-minY)*i/innerHeight) + minY
		text := fmt.Sprintf("%v", r)
		addLabel(gc, colors[8], text, leftMargin-float64(8*len(text)+16), height-bottomMargin-float64(i*40+16))
	}

	// https://github.com/llgcode/draw2d/blob/master/samples/line/line.go plot line to graph
	c := 0
	fMaxY := float64(maxY) - float64(diff)/float64(maxY-minY)
	for nodeID, pointList := range graphData {
		var colorID = (c + 1) % len(colors)
		// set color for line
		gc.SetFillColor(colors[colorID])
		gc.SetStrokeColor(colors[colorID])
		gc.SetFontSize(8)
		for i := 0; i < len(pointList)-1; i++ {
			p1 := pointList[i]
			p2 := pointList[i+1]
			log.Logger.Trace(p1, p2)

			xdiff := (width - (leftMargin + rightMargin)) / float64(maxX-minX)
			ydiff := (height - (topMargin + bottomMargin)) / float64(maxY-minY)
			// draw line
			drawLine(
				gc,
				float64(p1.X.Unix()-minX)*xdiff,
				ydiff*(fMaxY-p1.Y),
				float64(p2.X.Unix()-minX)*xdiff,
				ydiff*(fMaxY-p2.Y),
			)

		}
		// Add node right side label
		nodeName := *nodeIDToName[nodeID]
		if nodeName == "" {
			nodeName = fmt.Sprintf("Node %v", c)
		}
		addLabel(gc, colors[colorID], nodeName, width-rightMargin+16, topMargin+float64(c*20))

		c++
	}

	// saving png file
	os.Mkdir("graphs", os.ModePerm)
	fileName := fmt.Sprintf("graphs/%s-%v-%v.png", title, tgID, time.Now().Unix())
	// fileName := "graphs/a.png"
	draw2dimg.SaveToPngFile(fileName, img)

	// send message to the user
	fileable := tgbotapi.NewDocumentUpload(tgID, fileName)
	TgBot.Send(fileable)

	return nil
}

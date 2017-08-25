// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/lxn/walk"
	"github.com/russross/blackfriday"

	. "github.com/lxn/walk/declarative"
)

func main() {

	//main
	mw := &MyMainWindow{}

	if err := (MainWindow{
		Icon:     "img/search.ico",
		AssignTo: &mw.MainWindow,
		Title:    "BolgSearch -- by tim_zhang",

		MenuItems: []MenuItem{
			Menu{
				Text: "&编辑",
				Items: []MenuItem{
					Separator{},
					Action{
						Text:        "退出",
						OnTriggered: func() { mw.Close() },
					},
				},
			},
			Menu{
				Text: "&收藏",
				Items: []MenuItem{
					Separator{},
					Action{
						Text:        "查看",
						OnTriggered: mw.search,
					},
				},
			},
			Menu{
				Text: "&帮助",
				Items: []MenuItem{
					Action{
						Text:        "关于",
						OnTriggered: mw.aboutAction_Triggered,
					},
				},
			},
		},
		MinSize: Size{1000, 600},
		Layout:  VBox{MarginsZero: true},

		Children: []Widget{
			Composite{

				MaxSize: Size{0, 50},
				Layout:  HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &mw.favorites,
						Text:     "收藏",
					},
					PushButton{
						AssignTo: &mw.unfavorites,
						Text:     "取消收藏",
					},
					Label{Text: "网站: "},
					RadioButtonGroup{

						DataMember: "Web",
						Buttons: []RadioButton{
							RadioButton{
								Name:     "all",
								Text:     "全部",
								Value:    "1",
								AssignTo: &mw.all,
							},
							RadioButton{
								Name:     "jianshu",
								Text:     "简书",
								Value:    "2",
								AssignTo: &mw.jianshu,
							},
							RadioButton{
								Name:     "juejin",
								Text:     "掘金",
								Value:    "3",
								AssignTo: &mw.juejin,
							},
							RadioButton{
								Name:     "bokeyuan",
								Text:     "博客园",
								Value:    "4",
								AssignTo: &mw.bokeyuan,
							},
							RadioButton{
								Name:     "csdn",
								Text:     "CSDN",
								Value:    "5",
								AssignTo: &mw.csdn,
							},
							RadioButton{
								Name:     "oschina",
								Text:     "OSCHINA",
								Value:    "6",
								AssignTo: &mw.os,
							},
						},
					},
					Label{Text: "关键词: "},
					LineEdit{
						AssignTo: &mw.keywords,
						Text:     "golang",
					},
					PushButton{
						AssignTo: &mw.query,
						Text:     "查询",
					},
					PushButton{
						AssignTo: &mw.nextquery,
						Text:     "下一页",
					},
					PushButton{
						AssignTo: &mw.prvquery,
						Text:     "上一页",
					},
				},
			},

			Composite{
				Layout: Grid{Columns: 2, Spacing: 10},
				Children: []Widget{
					ListBox{
						MaxSize:               Size{200, 0},
						AssignTo:              &mw.lb,
						OnCurrentIndexChanged: mw.lb_CurrentIndexChanged,
						OnItemActivated:       mw.lb_ItemActivated,
					},

					WebView{
						//MinSize:  Size{1000, 0},
						AssignTo: &mw.wv,
					},
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	mw.query.Clicked().Attach(func() {
		go func() {

			mw.GetList(1)

		}()
	})

	mw.nextquery.Clicked().Attach(func() {
		go func() {

			mw.page = mw.page + 1
			mw.GetList(mw.page)

		}()
	})

	mw.prvquery.Clicked().Attach(func() {
		go func() {
			mw.page = mw.page - 1
			if mw.page == 0 {
				mw.page = 1
			}

			mw.GetList(mw.page)

		}()
	})

	mw.favorites.Clicked().Attach(func() {
		go func() {
			mw.addFavorite()
			mw.favorites.SetEnabled(false)
			mw.unfavorites.SetEnabled(true)
			walk.MsgBox(mw, "提示", "收藏成功", walk.MsgBoxIconInformation)
		}()
	})

	mw.unfavorites.Clicked().Attach(func() {
		go func() {
			mw.delFavorite()
			mw.favorites.SetEnabled(true)
			mw.unfavorites.SetEnabled(false)
			walk.MsgBox(mw, "提示", "取消收藏", walk.MsgBoxIconWarning)

		}()
	})

	//初始化
	mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/init.html")
	mw.all.SetChecked(true)
	mw.favorites.SetEnabled(false)
	mw.unfavorites.SetEnabled(false)
	mw.curtitle = ""
	mw.cururl = ""
	mw.Run()
}

type MyMainWindow struct {
	*walk.MainWindow
	lb          *walk.ListBox
	te          *walk.TextEdit
	wv          *walk.WebView
	all         *walk.RadioButton
	jianshu     *walk.RadioButton
	juejin      *walk.RadioButton
	bokeyuan    *walk.RadioButton
	csdn        *walk.RadioButton
	os          *walk.RadioButton
	keywords    *walk.LineEdit
	model       *Model
	query       *walk.PushButton
	nextquery   *walk.PushButton
	prvquery    *walk.PushButton
	favorites   *walk.PushButton
	unfavorites *walk.PushButton
	page        int
	curtitle    string
	cururl      string
}

type JianshuList struct {
	Q          string `json:"q"`
	Page       int    `json:"page"`
	Type       string `json:"type"`
	TotalCount int    `json:"total_count"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
	OrderBy    string `json:"order_by"`
	Entries    []struct {
		ID      int    `json:"id"`
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Content string `json:"content"`
		User    struct {
			ID        int    `json:"id"`
			Nickname  string `json:"nickname"`
			Slug      string `json:"slug"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		Notebook struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"notebook"`
		Commentable         bool      `json:"commentable"`
		PublicCommentsCount int       `json:"public_comments_count"`
		LikesCount          int       `json:"likes_count"`
		ViewsCount          int       `json:"views_count"`
		TotalRewardsCount   int       `json:"total_rewards_count"`
		FirstSharedAt       time.Time `json:"first_shared_at"`
	} `json:"entries"`
	RelatedUsers []struct {
		ID              int    `json:"id"`
		AvatarURL       string `json:"avatar_url"`
		Nickname        string `json:"nickname"`
		Slug            string `json:"slug"`
		TotalWordage    int    `json:"total_wordage"`
		TotalLikesCount int    `json:"total_likes_count"`
	} `json:"related_users"`
	MoreRelatedUsers   bool `json:"more_related_users"`
	RelatedCollections []struct {
		ID               int    `json:"id"`
		Title            string `json:"title"`
		Slug             string `json:"slug"`
		ImageURL         string `json:"image_url"`
		PublicNotesCount int    `json:"public_notes_count"`
		LikesCount       int    `json:"likes_count"`
	} `json:"related_collections"`
	MoreRelatedCollections bool `json:"more_related_collections"`
}

type JuejinList struct {
	D []struct {
		CollectionCount int    `json:"collectionCount"`
		OriginalURL     string `json:"originalUrl"`
		CommentsCount   int    `json:"commentsCount"`
		User            struct {
			ObjectID string `json:"objectId"`
			JobTitle string `json:"jobTitle"`
			Username string `json:"username"`
		} `json:"user"`
		ObjectID  string        `json:"objectId"`
		Content   string        `json:"content"`
		Title     string        `json:"title"`
		CreatedAt time.Time     `json:"createdAt"`
		Tags      []interface{} `json:"tags"`
		UpdatedAt time.Time     `json:"updatedAt"`
	} `json:"d"`
	M string `json:"m"`
	S int    `json:"s"`
}

type Item struct {
	name  string
	value string
}

type Items struct {
	items []Item
}

type Model struct {
	walk.ListModelBase
	items []Item
}

func (m *Model) ItemCount() int {
	return len(m.items)
}

func (m *Model) Value(index int) interface{} {
	return m.items[index].name
}

func (mw *MyMainWindow) aboutAction_Triggered() {
	walk.MsgBox(mw, "关于", "BlogSearch阅读器v1.0\n作者：tim_zhang 完成时间：2017-8-25 日", walk.MsgBoxIconQuestion)
}

func (mw *MyMainWindow) search() {
	mw.GetList(-999)
}

func (mw *MyMainWindow) GetList(page int) {
	mw.page = page
	keywords := mw.keywords.Text()
	enkeywords := url.QueryEscape(keywords)
	//mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/init.html")
	defer func() {
		mw.lb.SetCurrentIndex(-1)

	}()
	if page == -999 {
		go func() {
			rf := mw.readFavorite()
			m := &Model{items: rf.items}
			mw.lb.SetModel(m)
			mw.model = m
		}()
		return
	}
	if mw.all.Checked() == true {
		go func() {
			jianshu := getjianshList(enkeywords, strconv.Itoa(page))
			juejin := getjuejinList(enkeywords, strconv.Itoa(page))
			bokeyuan := getbokeyuanList(enkeywords, strconv.Itoa(page))
			csdn := getcsdnList(enkeywords, strconv.Itoa(page))
			os := getosList(enkeywords, strconv.Itoa(page))
			var k = 0
			var lens = 0
			if len(jianshu.items) > 0 {
				lens += len(jianshu.items)
			}
			if len(juejin.items) > 0 {
				lens += len(juejin.items)
			}
			if len(bokeyuan.items) > 0 {
				lens += len(bokeyuan.items)
			}
			if len(csdn.items) > 0 {
				lens += len(csdn.items)
			}
			if len(os.items) > 0 {
				lens += len(os.items)
			}
			m := &Model{items: make([]Item, lens)}

			if len(jianshu.items) > 0 {

				for _, jian := range jianshu.items {
					m.items[k] = jian
					k++
				}
			}
			if len(juejin.items) > 0 {

				for _, jue := range juejin.items {
					m.items[k] = jue
					k++
				}
			}
			if len(bokeyuan.items) > 0 {

				for _, bo := range bokeyuan.items {
					m.items[k] = bo
					k++
				}
			}
			if len(csdn.items) > 0 {

				for _, cs := range csdn.items {
					m.items[k] = cs
					k++
				}
			}
			if len(os.items) > 0 {

				for _, osl := range os.items {
					m.items[k] = osl
					k++
				}
			}
			mw.lb.SetModel(m)
			mw.model = m
		}()
		return
	}
	if mw.jianshu.Checked() == true {
		go func() {
			jianshu := getjianshList(enkeywords, strconv.Itoa(page))
			m := &Model{items: jianshu.items}
			mw.lb.SetModel(m)
			mw.model = m
		}()
		return
	}
	if mw.juejin.Checked() == true {
		go func() {
			juejin := getjuejinList(enkeywords, strconv.Itoa(page))
			m := &Model{items: juejin.items}
			mw.lb.SetModel(m)
			mw.model = m
		}()
		return
	}
	if mw.bokeyuan.Checked() == true {
		go func() {
			bokeyuan := getbokeyuanList(enkeywords, strconv.Itoa(page))
			m := &Model{items: bokeyuan.items}
			mw.lb.SetModel(m)
			mw.model = m
		}()
		return
	}
	if mw.csdn.Checked() == true {
		go func() {
			csdn := getcsdnList(enkeywords, strconv.Itoa(page))
			m := &Model{items: csdn.items}
			mw.lb.SetModel(m)
			mw.model = m
		}()
		return
	}
	if mw.os.Checked() == true {
		go func() {
			os := getosList(enkeywords, strconv.Itoa(page))
			m := &Model{items: os.items}
			mw.lb.SetModel(m)
			mw.model = m
		}()
		return
	}

	return
}

func (mw *MyMainWindow) lb_CurrentIndexChanged() {

	i := mw.lb.CurrentIndex()
	if i < 0 {
		return
	}
	//defer mw.lb.SetCurrentIndex(-1)
	item := &mw.model.items[i]
	s := strings.Split(item.value, "|")
	if len(s) == 2 {
		switch s[0] {
		case "jianshu":
			mw.getjianshu(item.name, s[1])
		case "juejin":
			mw.getjuejin(item.name, s[1])
		case "bokeyuan":
			mw.getbokeyuan(item.name, s[1])
		case "csdn":
			mw.getcsdn(item.name, s[1])
		case "os":
			mw.getos(item.name, s[1])
		}
		//记录当前页面的标题和链接
		mw.curtitle = item.name
		mw.cururl = item.value
		if mw.checkFavorite() {
			mw.favorites.SetEnabled(false)
			mw.unfavorites.SetEnabled(true)
		} else {
			mw.favorites.SetEnabled(true)
			mw.unfavorites.SetEnabled(false)
		}
	} else {
		mw.curtitle = ""
		mw.cururl = ""
		mw.favorites.SetEnabled(false)
		mw.unfavorites.SetEnabled(false)
		mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/xxx.html")
	}

	return
}

func (mw *MyMainWindow) lb_ItemActivated() {
	mw.curtitle = ""
	mw.cururl = ""
	mw.favorites.SetEnabled(false)
	mw.unfavorites.SetEnabled(false)
	mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/xxx.html")
}

func (mw *MyMainWindow) getjianshu(name string, value string) {
	go func() {
		re, _ := regexp.Compile("<img[\\S\\s]+?>")
		userFile := "data/xxx.html"
		fout, err := os.Create(userFile)
		defer fout.Close()
		if err != nil {

			return
		}
		//walk.MsgBox(mw, "查询", "http://www.jianshu.com/p/"+value, walk.MsgBoxIconInformation)
		doc, err := goquery.NewDocument("http://www.jianshu.com/p/" + value)
		if err != nil {
			fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + "————简书</h3><a href='http://www.jianshu.com/p/" + value + "' target='_blank'>文章地址</a>")
		}
		ht, _ := doc.Find(".show-content").Html()

		fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + "————简书</h3><a href='http://www.jianshu.com/p/" + value + "' target='_blank'>文章地址</a>" + re.ReplaceAllString(ht, ""))

		mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/xxx.html")
	}()

}

func (mw *MyMainWindow) getjuejin(name string, value string) {
	go func() {
		userFile := "data/xxx.html"
		fout, err := os.Create(userFile)
		defer fout.Close()
		if err != nil {
			return
		}
		//walk.MsgBox(mw, "查询", "https://juejin.im/entry/"+value, walk.MsgBoxIconInformation)
		doc, err := goquery.NewDocument("https://juejin.im/entry/" + value)
		if err != nil {
			//log.Fatal(err)
			fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + " ————掘金</h3><a href='https://juejin.im/entry/" + value + "' target='_blank'>文章地址</a>")
		}
		ht, _ := doc.Find(".entry-content").Html()
		re, _ := regexp.Compile("<img[\\S\\s]+?>")

		fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + "————掘金</h3><a href='https://juejin.im/entry/" + value + "' target='_blank'>文章地址</a>" + re.ReplaceAllString(ht, ""))
		mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/xxx.html")
	}()

}

func (mw *MyMainWindow) getbokeyuan(name string, value string) {
	go func() {
		userFile := "data/xxx.html"
		fout, err := os.Create(userFile)
		defer fout.Close()
		if err != nil {
			return
		}
		//walk.MsgBox(mw, "查询", value, walk.MsgBoxIconInformation)
		doc, err := goquery.NewDocument(value)
		if err != nil {
			//log.Fatal(err)
			fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + " ————博客园</h3><a href='" + value + "' target='_blank'>文章地址</a>")
		}
		ht, _ := doc.Find("#cnblogs_post_body").Html()
		re, _ := regexp.Compile("<img[\\S\\s]+?>")

		fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + "————博客园</h3><a href='" + value + "' target='_blank'>文章地址</a>" + re.ReplaceAllString(ht, ""))
		mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/xxx.html")
	}()

}

func (mw *MyMainWindow) getcsdn(name string, value string) {
	go func() {
		userFile := "data/xxx.html"
		fout, err := os.Create(userFile)
		defer fout.Close()
		if err != nil {
			return
		}
		//walk.MsgBox(mw, "查询", value, walk.MsgBoxIconInformation)
		doc, err := goquery.NewDocument(value)
		if err != nil {
			//log.Fatal(err)
			fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + " ————CSDN</h3><a href='" + value + "' target='_blank'>文章地址</a>")
		}
		ht, _ := doc.Find("#article_content").Html()
		re, _ := regexp.Compile("<img[\\S\\s]+?>")
		hs := re.ReplaceAllString(ht, "")
		res, _ := regexp.Compile("<script[\\S\\s]+?>")

		fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + "————CSDN</h3><a href='" + value + "' target='_blank'>文章地址</a>" + res.ReplaceAllString(hs, ""))
		mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/xxx.html")
	}()

}

func (mw *MyMainWindow) getos(name string, value string) {
	go func() {
		userFile := "data/xxx.html"
		fout, err := os.Create(userFile)
		defer fout.Close()
		if err != nil {
			return
		}
		//walk.MsgBox(mw, "查询", value, walk.MsgBoxIconInformation)
		doc, err := goquery.NewDocument(value)
		if err != nil {
			//log.Fatal(err)
			fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + " ————OSCHINA</h3><a href='" + value + "' target='_blank'>文章地址</a>")
		}
		ht, _ := doc.Find(".blog-body").Html()
		hs, _ := doc.Find(".noshow_content").Html()
		if len(hs) != 0 {
			ht = string(blackfriday.MarkdownBasic([]byte(hs)))

		}
		re, _ := regexp.Compile("<img[\\S\\s]+?>")

		fout.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><h3>" + name + "————OSCHINA</h3><a href='" + value + "' target='_blank'>文章地址</a>" + re.ReplaceAllString(ht, ""))
		mw.wv.SetURL("file:///" + getCurrentDirectory() + "/data/xxx.html")
	}()
}

func httpDo(method string, url string) (bodys string) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {

	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36")
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	bodys = string(body)

	if err != nil {
		// handle error
		return ""
	}

	// fmt.Println(string(body))
	return bodys
}

func (mw *MyMainWindow) addFavorite() {
	mw.delFavorite()
	rf := mw.readFavorite()
	st := mw.curtitle + "<$>" + mw.cururl + "$$"
	for i, ite := range rf.items {
		if i == len(rf.items)-1 {

			st = st + ite.name + "<$>" + ite.value + "$$"
		} else {
			st = st + ite.name + "<$>" + ite.value

		}
	}

	userFile := "data/favorite"
	fout, err := os.Create(userFile)
	defer fout.Close()
	if err != nil {
		return
	}
	fout.WriteString(st)
}

func (mw *MyMainWindow) delFavorite() {

	rf := mw.readFavorite()
	sb := mw.curtitle + "<$>" + mw.cururl
	st := ""
	for i, ite := range rf.items {
		if sb != ite.name+"<$>"+ite.value {
			if i == 0 {
				st = st + ite.name + "<$>" + ite.value

			} else {

				st = st + "$$" + ite.name + "<$>" + ite.value
			}
		}
	}

	userFile := "data/favorite"
	fout, err := os.Create(userFile)
	defer fout.Close()
	if err != nil {
		return
	}
	fout.WriteString(st)
}

func (mw *MyMainWindow) readFavorite() Items {
	f, _ := os.Open("data/favorite")
	fd, _ := ioutil.ReadAll(f)
	fs := string(fd)
	s := strings.Split(fs, "$$")
	if len(s[0]) > 0 {
		data := Items{items: make([]Item, len(s))}
		for i, ite := range s {
			if len(ite) > 0 {
				b := strings.Split(ite, "<$>")
				data.items[i].name = b[0]
				data.items[i].value = b[1]
			}

		}
		return data
	} else {
		data := Items{}
		return data
	}
}

func (mw *MyMainWindow) checkFavorite() bool {
	rf := mw.readFavorite()
	sb := mw.curtitle + "<$>" + mw.cururl
	st := false
	for _, ite := range rf.items {
		if sb == ite.name+"<$>"+ite.value {
			st = true
		}
	}
	return st
}

func getjianshList(keywords string, page string) Items {
	body := httpDo("GET", "http://www.jianshu.com/search/do?q="+keywords+"&type=note&page="+page+"&order_by=default")
	var r JianshuList
	json.Unmarshal([]byte(body), &r)
	data := Items{items: make([]Item, len(r.Entries))}
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	for i, en := range r.Entries {
		data.items[i] = Item{re.ReplaceAllString(en.Title, ""), re.ReplaceAllString("jianshu|"+en.Slug, "")}
	}

	return data
}

func getjuejinList(keywords string, page string) Items {
	body := httpDo("GET", "https://search-merger-ms.juejin.im/v1/search?query="+keywords+"&page="+page+"&raw_result=false&src=web")
	var r JuejinList
	json.Unmarshal([]byte(body), &r)
	data := Items{items: make([]Item, len(r.D))}
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	for i, en := range r.D {
		data.items[i] = Item{re.ReplaceAllString(en.Title, ""), re.ReplaceAllString("juejin|"+en.ObjectID, "")}
	}

	return data
}

func getbokeyuanList(keywords string, page string) Items {
	data := Items{items: make([]Item, 15)}
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	doc, _ := goquery.NewDocument("http://zzk.cnblogs.com/s/blogpost?Keywords=" + keywords + "&pageindex=" + page)
	doc.Find(".searchItemTitle").Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		value, _ := s.Find("a").Attr("href")
		data.items[i] = Item{strings.Replace(re.ReplaceAllString(title, ""), " ", "", -1), "bokeyuan|" + value}
	})
	return data
}

func getcsdnList(keywords string, page string) Items {
	data := Items{items: make([]Item, 10)}
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	doc, _ := goquery.NewDocument("http://so.csdn.net/so/search/s.do?p=" + page + "&q=" + keywords + "&t=blog&domain=&o=&s=&u=null&l=null&f=null")
	doc.Find(".search-list").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a").Text()
		value, _ := s.Find("a").Attr("href")
		data.items[i] = Item{re.ReplaceAllString(title, ""), "csdn|" + value}
	})
	return data
}

func getosList(keywords string, page string) Items {
	data := Items{items: make([]Item, 20)}
	body := httpDo("GET", "https://www.oschina.net/search?scope=blog&sort_by_time=1&q="+keywords+"&p="+page)
	digitsRegexp := regexp.MustCompile(`"(.*)"`)
	s := digitsRegexp.FindStringSubmatch(body)
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	doc, _ := goquery.NewDocument("https://www.oschina.net/search" + s[1])
	doc.Find(".obj_type_3").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a").Text()
		value, _ := s.Find("a").Attr("href")
		data.items[i] = Item{re.ReplaceAllString(title, ""), "os|" + value}
	})
	return data
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type MolyBook struct {
	Author    string
	Title     string
	Subtitle  string
	MolyUrl   string
	CoverPath string
}

func getCSRFToken(url string, cookies *[]*http.Cookie) string {
	resp, respCookies := performHTTPRequest("GET", url, nil, *cookies)
	updateCookieJar(cookies, &respCookies)

	//<meta name="csrf-token" content="DWtRvc2CXNim1HdyyLRnw3obF9J3am7c2JfYDj1lI1q6oluvTeI+oiIvkdVYSakF0h0LJIlbGJhQ3DMdInLzbA==" />
	csrfRegex, _ := regexp.Compile(`<meta name="csrf-token" content="[A-Za-z\d+=/]+" />`)
	csrfTag := csrfRegex.FindString(string(resp))
	csrfToken := strings.Split(csrfTag, `"`)[3]

	return csrfToken
}

func Login(username string, password string) []*http.Cookie {

	fmt.Println("[+] Logging in as ", username)

	//get a CSRF token
	var session []*http.Cookie
	csrfToken := getCSRFToken("https://moly.hu/belepes", &session)

	//login
	loginBody := url.Values{
		"utf8":                      {"✓"},
		"authenticity_token":        {csrfToken},
		"user_session[email]":       {username},
		"user_session[password]":    {password},
		"user_session[remember_me]": {"0"},
		"commit":                    {"Belépés"},
	}

	resp, respCookies := performHTTPRequest("POST", "https://moly.hu/azonositas", []byte(loginBody.Encode()), session)
	updateCookieJar(&session, &respCookies)

	//test if cookie is valid
	if !strings.Contains(string(resp), `You are being <a href="https://moly.hu/">`) {
		panic("[!] Moly Login failed!")
	}

	fmt.Println("[+] Moly Login successful!")

	return session
}

func NewBook(book *MolyBook, session []*http.Cookie) string {
	fmt.Println("[+] Adding book ", book.Title)
	csrfToken := getCSRFToken("https://moly.hu/konyvek/uj", &session)

	loginBody := url.Values{
		"utf8":               {"✓"},
		"authenticity_token": {csrfToken},
		"book[author]":       {book.Author},
		"book[title]":        {book.Title},
		"book[subtitle]":     {book.Subtitle},
		"commit":             {"Mentés"},
	}

	resp, respCookies := performHTTPRequest("POST", "https://moly.hu/konyvek", []byte(loginBody.Encode()), session)
	updateCookieJar(&session, &respCookies)

	if strings.Contains(string(resp), "body>You are being <a href=") {
		bookUrl := strings.Split(string(resp), `"`)[1]
		fmt.Println("[+] Book created @", bookUrl)
		return string(bookUrl)
	}

	panic("[!] Book creation probably failed, check on the site")
	return ""
}

func uploadCover(molyUrl string, coverPath string, session []*http.Cookie) {

	fmt.Println("[+] Adding book cover", coverPath)
	csrfToken := getCSRFToken(molyUrl+"/boritok/uj", &session)

	loginBody := map[string]io.Reader{
		"utf8":                 strings.NewReader("✓"),
		"commit":               strings.NewReader("Mentés"),
		"remotipart_submitted": strings.NewReader("true"),
		"authenticity_token":   strings.NewReader(csrfToken),
		"X-Requested-With":     strings.NewReader("IFrame"),
		"X-HTTP-Accept":        strings.NewReader("text/javascript, application/javascript, application/ecmascript, application/x-ecmascript, */*; q=0.01"),
		"cover[image]":         mustOpen(coverPath),
	}

	resp, respCookies := performMultiPartForm(molyUrl+"/boritok", loginBody, session)
	updateCookieJar(&session, &respCookies)

	if !strings.Contains(string(resp), "modalbox.hide();") {
		fmt.Println("[+] Cover upload successful")
	} else {
		fmt.Println("[+] Cover upload failed")

	}
}

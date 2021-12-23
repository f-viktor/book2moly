package main

func main() {
	args := parseArgs()

	book := MolyBook{
		Author:   "Example Author",
		Title:    "Example Book",
		Subtitle: "Subtitle",
	}
	molyCookies := Login(args.Username, args.Password) //get session cookie
	book.MolyUrl = NewBook(&book, molyCookies)         // get csrf token post book

	book.CoverPath = "cover.png"
	if book.CoverPath != "" {
		uploadCover(book.MolyUrl, book.CoverPath, molyCookies)
	}
}

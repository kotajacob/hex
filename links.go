package main

import "strings"

func newLinkReplacer(domain string) *strings.Replacer {
	linkReplacer := strings.NewReplacer(
		"https://hexbear.net/communities",
		domain+"/communities",
		"https://www.hexbear.net/communities",
		domain+"/communities",
		"https://hexbear.net/ppb",
		domain+"/ppb",
		"https://www.hexbear.net/ppb",
		domain+"/ppb",
		"https://hexbear.net/post/",
		domain+"/post/",
		"https://www.hexbear.net/post/",
		domain+"/post/",
		"https://hexbear.net/c/",
		domain+"/c/",
		"https://www.hexbear.net/c/",
		domain+"/c/",
		"https://hexbear.net/u/",
		domain+"/u/",
		"https://www.hexbear.net/u/",
		domain+"/u/",
	)
	return linkReplacer
}

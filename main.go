// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/chrome/service.py
// http://code.google.com/p/selenium/wiki/ChromeDriver

package main

import (
	"flag"
	"fmt"
	"runtime"
)

func main() {
	if runtime.GOOS != "linux" {
		fmt.Println(fmt.Errorf("selenium not yet implemented on %s", runtime.GOOS))
		return
	}
	if runtime.GOARCH != "amd64" {
		fmt.Println(fmt.Errorf("selenium not yet implemented on %s", runtime.GOARCH))
		return
	}
	var browser_name string
	flag.StringVar(&browser_name, "browser", "chromium", "-browser=chromium|chrome|firefox|iceweasel|ie|opera")
	flag.Parse()
	if browser_name != "chromium" {
		fmt.Println(fmt.Errorf("selenium not yet implemented on %s", browser_name))
		return
	}
	fmt.Println("browser ==", browser_name)
}

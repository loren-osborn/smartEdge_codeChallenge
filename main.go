// codechallenge is a dummy program used to demonstrate proper docker usage
package codechallenge

import "fmt"

// DoStuff is the program entry-point with all dependencies injected. This
// allows us to test respecting public vs private methods by moving it outside
// the "main" package.
func DoStuff() {
	fmt.Println(`{
    "message":"theAnswerIs42",
    "signature":"MGUCMCDwlFyVdD620p0hRLtABoJTR7UNgwj8g2r0ipNbWPi4Us57YfxtSQJ3dAkHslyBbwIxAKorQmpWl9QdlBUtACcZm4kEXfL37lJ+gZ/hANcTyuiTgmwcEC0FvEXY35u2bKFwhA==",
    "pubkey":"-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEI5/0zKsIzou9hL3ZdjkvBeVZFKpDwxTb\nfiDVjHpJdu3+qOuaKYgsLLiO9TFfupMYHLa20IqgbJSIv/wjxANH68aewV1q2Wn6\nvLA3yg2mOTa/OHAZEiEf7bVEbnAov+6D\n-----END PUBLIC KEY-----\n"
}`)
}

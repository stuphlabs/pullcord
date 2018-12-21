package authentication

import (
	"github.com/stretchr/testify/assert"
	// "github.com/stuphlabs/pullcord"
	"net/http"
	"testing"
)

// TestMinSessionHandlerFirstPass tests if a MinSessionHandler will even give an
// initial cookie.
//
// Steps:
// 	1. Create a new MinSessionHandler to test.
// 	2. Run the cookie mask with an empty list for the input cookies.
// 	3. Verify that we received a cookie.
func TestMinSessionHandlerFirstPass(t *testing.T) {
	/* setup */

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh, err := handler.GetSession()
	assert.NoError(t, err)
	fwd, stc, err := sesh.CookieMask(nil)

	/* check */
	assert.NoError(t, err)
	assert.Nil(t, fwd)
	assert.NotNil(t, stc)
	if stc != nil {
		assert.Equal(t, 1, len(stc))
	}
	assert.NotNil(t, sesh)
	/*
		if sesh != nil {
			assert.Equal(t, 0, len(sesh))
		}
	*/
}

// TestMinSessionHandlerReuseCookie tests if a MinSessionHandler will accept a
// cookie it just gave us.
//
// Steps:
// 	1. Create a new MinSessionHandler to test.
// 	2. Run the cookie mask with an empty list for the input cookies.
// 	3. Run the cookie mask again, this time including the cookie we just
//	   received in the input cookie list.
// 	4. Verify that we did not receive another cookie.
func TestMinSessionHandlerReuseCookie(t *testing.T) {
	/* setup */
	var localCookies []*http.Cookie

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	for _, cookie := range stc1 {
		localCookies = append(localCookies, cookie)
	}
	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(localCookies)

	/* check */
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Nil(t, fwd1)
	assert.Nil(t, fwd2)
	assert.NotNil(t, stc1)
	if stc1 != nil {
		assert.Equal(t, 1, len(stc1))
	}
	assert.Nil(t, stc2)
	assert.NotNil(t, sesh1)
	/*
		if sesh1 != nil {
			assert.Equal(t, 0, len(sesh1))
		}
	*/
	assert.NotNil(t, sesh2)
	/*
		if sesh2 != nil {
			assert.Equal(t, 0, len(sesh2))
		}
	*/
}

// TestMinSessionHandlerSessionDataPreservation tests if a MinSessionHandler
// preserves session data between requests.
//
// Steps:
// 	1. Create a MinSessionHandler to test.
// 	2. Run the cookie mask to get a new cookie and session.
// 	3. Insert a new entry into the session data.
// 	4. Run the cookie mask again with the same cookie we received.
// 	5. Verify that the new session contains the same data we added to the
//	   previous session.
func TestMinSessionHandlerSessionDataPreservation(t *testing.T) {
	/* setup */
	var localCookies []*http.Cookie
	expectedData := make(map[string]interface{})
	expectedKey := "test key"

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	_, present1 := sesh1.GetValue(expectedKey)
	for _, cookie := range stc1 {
		localCookies = append(localCookies, cookie)

		/* intermediate check */
		_, present1 = sesh1.GetValue(expectedKey)
		assert.Error(t, present1)
		assert.Equal(t, present1, NoSuchSessionValueError)

		expectedString := "saving data into " + cookie.Name + " cookie"
		expectedData[expectedKey] = expectedString
		err = sesh1.SetValue(expectedKey, expectedString)
		assert.NoError(t, err)
	}
	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(localCookies)

	/* check */
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Nil(t, fwd1)
	assert.Nil(t, fwd2)
	assert.NotNil(t, stc1)
	if stc1 != nil {
		assert.Equal(t, 1, len(stc1))
	}
	assert.Nil(t, stc2)
	assert.NotNil(t, sesh1)
	assert.NotNil(t, sesh2)
	/*
		if sesh2 != nil {
			assert.Equal(t, 1, len(sesh2))
		}
	*/

	assert.Error(t, present1)
	assert.Equal(t, present1, NoSuchSessionValueError)

	actualData2, present2 := sesh2.GetValue(expectedKey)
	assert.NoError(t, present2)
	assert.NotEqual(t, present2, NoSuchSessionValueError)
	assert.Equal(t, expectedData[expectedKey], actualData2)
}

// TestMinSessionHandlerBadCookie tests if a MinSessionHandler recognizes a bad
// cookie.
//
// Steps:
// 	1. Create a new MinSessionHandler to test.
// 	2. Run the cookie mask in order to get a good cookie.
// 	3. Tamper with the cookie.
// 	4. Run the cookie mask with the tampered cookie.
// 	5. Verify that the bad cookie was rejected and replaced by another good
//	   cookie.
func TestMinSessionHandlerBadCookie(t *testing.T) {
	/* setup */
	var localCookies []*http.Cookie
	var badCookie http.Cookie

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	for _, cookie := range stc1 {
		cookie.Value = cookie.Value + "bad"
		badCookie.Name = cookie.Name
		badCookie.Value = cookie.Value
		localCookies = append(localCookies, cookie)
	}
	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(localCookies)

	/* check */
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Nil(t, fwd1)
	assert.Nil(t, fwd2)
	assert.NotNil(t, stc1)
	if stc1 != nil {
		assert.Equal(t, 1, len(stc1))
	}
	assert.NotNil(t, stc2)
	if stc2 != nil {
		assert.Equal(t, 2, len(stc2))
	}
	badCookieDeleted := false
	for _, cookie := range stc2 {
		if cookie.Name == badCookie.Name {
			assert.Equal(t, badCookie.Value, cookie.Value)
			assert.Equal(t, -1, cookie.MaxAge)
			badCookieDeleted = true
		}
	}
	assert.True(t, badCookieDeleted)
	assert.NotNil(t, sesh1)
	/*
		if sesh1 != nil {
			assert.Equal(t, 0, len(sesh1))
		}
	*/
	assert.NotNil(t, sesh2)
	/*
		if sesh2 != nil {
			assert.Equal(t, 0, len(sesh2))
		}
	*/
}

// TestMinSessionHandlerInvalidCookie tests if a MinSessionHandler rejects a
// cookie that it did not create.
//
// Steps:
// 	1. Forge a cookie that would match the MinSessionHandler's regular
//	   expression.
// 	2. Create a new MinSessionHandler to test that will create cookies with
//	   the same naming mechanism as our foged cookie.
// 	3. Run the cookie mask with the forged cookie.
// 	4. Verify that the forged cookie was rejected and replaced by another
//	   cookie.
func TestMinSessionHandlerInvalidCookie(t *testing.T) {
	/* setup */
	var invalidCookie http.Cookie
	var localCookies []*http.Cookie
	invalidCookie.Name = "testHandler-"
	for i := 0; i < minSessionCookieNameRandSize; i++ {
		invalidCookie.Name += "ff"
	}
	invalidCookie.Value = "foo"
	localCookies = append(localCookies, &invalidCookie)

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh, err := handler.GetSession()
	assert.NoError(t, err)
	fwd, stc, err := sesh.CookieMask(localCookies)

	/* check */
	assert.NoError(t, err)
	assert.Nil(t, fwd)
	assert.NotNil(t, stc)
	if stc != nil {
		assert.Equal(t, 2, len(stc))
	}
	badCookieDeleted := false
	for _, cookie := range stc {
		if cookie.Name == invalidCookie.Name {
			assert.Equal(t, invalidCookie.Value, cookie.Value)
			assert.Equal(t, -1, cookie.MaxAge)
			badCookieDeleted = true
		}
	}
	assert.True(t, badCookieDeleted)
	assert.NotNil(t, sesh)
	/*
		if sesh != nil {
			assert.Equal(t, 0, len(sesh))
		}
	*/
}

// TestMinSessionHandlerMultiSession tests if a MinSessionHandler can correctly
// track multiple sessions.
//
// Steps:
// 	 1. Create a new MinSessionHandler to test.
// 	 2. Run the cookie mask with an empty cookie list.
// 	 3. Save the cookie that we just received into cookie list 1.
// 	 4. Set a value in the session we just received.
// 	 5. Run the cookie mask with another empty cookie list.
// 	 6. Save the cookie that we just received into cookie list 2.
// 	 7. Set a value in the session we just received.
// 	 8. Run the cookie mask with cookie list 2.
// 	 9. Record what value was in the session we just received.
// 	10. Set a new value in the session we just received.
// 	11. Run the cookie mask with cookie list 1.
// 	12. Record what value was in the session we just received.
// 	13. Set a new value in the session we just received.
// 	14. Run the cookie mask with cookie list 2.
// 	15. Record what value was in the session we just received.
// 	16. Verify that session data was not present initially.
// 	17. Verify that the session data was what was expected for subsequent
//	    accesses with the same cookie.
func TestMinSessionHandlerMultiSession(t *testing.T) {
	/* setup */
	var (
		localCookies1    []*http.Cookie
		localCookies2    []*http.Cookie
		seshKey          = "test key"
		expectedPresent1 = NoSuchSessionValueError
		actualPresent1   error
		expectedPresent2 = NoSuchSessionValueError
		actualPresent2   error
		expectedPresent3 = error(nil)
		actualPresent3   error
		expectedPresent4 = error(nil)
		actualPresent4   error
		expectedPresent5 = error(nil)
		actualPresent5   error
		expectedValue3   = "test 3"
		actualValue3     interface{}
		expectedValue4   = "test 4"
		actualValue4     interface{}
		expectedValue5   = "test 5"
		actualValue5     interface{}
		saveValue1       = expectedValue4
		saveValue2       = expectedValue3
		saveValue3       = expectedValue5
	)

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")

	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(localCookies1)
	for _, cookie := range stc1 {
		localCookies1 = append(localCookies1, cookie)
	}
	_, actualPresent1 = sesh1.GetValue(seshKey)
	err = sesh1.SetValue(seshKey, saveValue1)
	assert.NoError(t, err)

	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(localCookies2)
	for _, cookie := range stc2 {
		localCookies2 = append(localCookies2, cookie)
	}
	_, actualPresent2 = sesh2.GetValue(seshKey)
	err = sesh2.SetValue(seshKey, saveValue2)
	assert.NoError(t, err)

	sesh3, err := handler.GetSession()
	assert.NoError(t, err)
	fwd3, stc3, err3 := sesh3.CookieMask(localCookies2)
	actualValue3, actualPresent3 = sesh3.GetValue(seshKey)
	err = sesh3.SetValue(seshKey, saveValue3)
	assert.NoError(t, err)

	sesh4, err := handler.GetSession()
	assert.NoError(t, err)
	fwd4, stc4, err4 := sesh4.CookieMask(localCookies1)
	actualValue4, actualPresent4 = sesh4.GetValue(seshKey)

	sesh5, err := handler.GetSession()
	assert.NoError(t, err)
	fwd5, stc5, err5 := sesh5.CookieMask(localCookies2)
	actualValue5, actualPresent5 = sesh5.GetValue(seshKey)

	/* check */
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NoError(t, err4)
	assert.NoError(t, err5)
	assert.Nil(t, fwd1)
	assert.Nil(t, fwd2)
	assert.Nil(t, fwd3)
	assert.Nil(t, fwd4)
	assert.Nil(t, fwd5)
	assert.NotNil(t, stc1)
	if stc1 != nil {
		assert.Equal(t, 1, len(stc1))
	}
	assert.NotNil(t, stc2)
	if stc2 != nil {
		assert.Equal(t, 1, len(stc2))
	}
	assert.Nil(t, stc3)
	assert.Nil(t, stc4)
	assert.Nil(t, stc5)
	assert.Equal(t, expectedPresent1, actualPresent1)
	assert.Equal(t, expectedPresent2, actualPresent2)
	assert.Equal(t, expectedPresent3, actualPresent3)
	assert.Equal(t, expectedPresent4, actualPresent4)
	assert.Equal(t, expectedPresent5, actualPresent5)
	assert.Equal(t, expectedValue3, actualValue3)
	assert.Equal(t, expectedValue4, actualValue4)
	assert.Equal(t, expectedValue5, actualValue5)
}

// TestMinSessionHandlerBadCookieDestroysSession tests if a MinSessionHandler
// destroys a session after a bad cookie.
//
// Steps:
// 	1. Create a MinSessionHandler to test.
// 	2. Run the cookie mask with an empty cookie list.
// 	3. Save the cookie we received into the good cookie list.
// 	4. Tamper with a copy of the cookie we received, and save this bad
//	   cookie into the bad cookie list.
// 	5. Set a value in the session we just received.
// 	6. Run the cookie mask with the bad cookie list.
// 	7. Run the cookie mask with the good cookie list.
// 	8. Verify that the subsequent sessions we received did not contain the
//	   value we had previously set.
// 	9. Verify that each time the provided cookie was rejected and we
//	   received a replacement cookie.
func TestMinSessionHandlerBadCookieDestroysSession(t *testing.T) {
	/* setup */
	var (
		goodCookies            []*http.Cookie
		badCookies             []*http.Cookie
		badCookie              http.Cookie
		seshKey                = "test key"
		expectedSeshPresent1   = NoSuchSessionValueError
		actualSeshPresent1     error
		expectedSeshPresent2   = NoSuchSessionValueError
		actualSeshPresent2     error
		expectedSeshPresent3   = NoSuchSessionValueError
		actualSeshPresent3     error
		expectedCookiePresent2 = true
		actualCookiePresent2   bool
		expectedCookiePresent3 = true
		actualCookiePresent3   bool
		saveValue              = "foo"
	)

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")

	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	for _, goodCookie := range stc1 {
		goodCookies = append(goodCookies, goodCookie)
		badCookie.Name = goodCookie.Name
		badCookie.Value = goodCookie.Value + " bar"
		badCookies = append(badCookies, &badCookie)
	}
	_, actualSeshPresent1 = sesh1.GetValue(seshKey)
	err = sesh1.SetValue(seshKey, saveValue)
	assert.NoError(t, err)

	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(badCookies)
	actualCookiePresent2 = false
	for _, cookie := range stc2 {
		if cookie.Name == badCookie.Name {
			actualCookiePresent2 = true
		}
	}
	_, actualSeshPresent2 = sesh2.GetValue(seshKey)

	sesh3, err := handler.GetSession()
	assert.NoError(t, err)
	fwd3, stc3, err3 := sesh3.CookieMask(goodCookies)
	actualCookiePresent3 = false
	for _, cookie := range stc3 {
		if cookie.Name == badCookie.Name {
			actualCookiePresent3 = true
		}
	}
	_, actualSeshPresent3 = sesh3.GetValue(seshKey)

	/* check */
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Nil(t, fwd1)
	assert.Nil(t, fwd2)
	assert.Nil(t, fwd3)
	assert.NotNil(t, stc1)
	if stc1 != nil {
		assert.Equal(t, 1, len(stc1))
	}
	assert.NotNil(t, stc2)
	if stc2 != nil {
		assert.Equal(t, 2, len(stc2))
	}
	assert.NotNil(t, stc3)
	if stc3 != nil {
		assert.Equal(t, 2, len(stc3))
	}
	assert.Equal(t, expectedSeshPresent1, actualSeshPresent1)
	assert.Equal(t, expectedSeshPresent2, actualSeshPresent2)
	assert.Equal(t, expectedSeshPresent3, actualSeshPresent3)
	assert.Equal(t, expectedCookiePresent2, actualCookiePresent2)
	assert.Equal(t, expectedCookiePresent3, actualCookiePresent3)
}

// TestMinSessionHandlerNonInterfering tests if two MinSessionHandlers interfere
// with one another.
//
// Steps:
// 	 1. Create two MinSessionHandlers to test with.
// 	 2. Run the cookie mask of the first MinSessionHendler with an empty
// 	    cookie list.
// 	 3. Save the cookie we just received into cookie list 1.
// 	 4. Tamper with a copy of the cookie we just received and place it into
// 	    cookie list 2.
// 	 5. Set a value in the session from this first MinSessionHandler.
// 	 6. Run the cookie mask of the second MinSessionHandler with cookie list
// 	    1.
// 	 7. Record what cookies are being forwarded.
// 	 8. Add the cookie we just received from the second MinSessionHandler
// 	    into cookie lists 1 and 2.
// 	 9. Set a value in the session from the second MinSessionHandler.
// 	10. Run the cookie mask of the second MinSessionHandler with cookie list
// 	    2.
// 	11. Record what cookies are being forwarded.
// 	12. Record the value in the session we just received.
// 	13. Run the cookie mask of the first MinSessionHandler with cookie list
// 	    2.
// 	14. Record what cookies are being forwarded.
// 	15. Look for the value in the session we just received.
// 	16. Verify that the cookies from each MinSessionHandler were being
// 	    properly forwarded by the opposite MinSessionHandler.
// 	17. Verify that the correct cookie was accepted by the second
// 	    MinSessionHandler.
// 	18. Verify that the session data was preserved by the second
// 	    MinSessionHandler.
// 	19. Verify that the tampered cookie was rejected by the first
// 	    MinSessionHandler.
// 	20. Verify that the session data was destroyed by the first
// 	    MinSessionHandler.
func TestMinSessionHandlerNonInterfering(t *testing.T) {
	/* setup */
	var (
		localCookies1    []*http.Cookie
		localCookies2    []*http.Cookie
		seshKey          = "test key"
		expectedPresent1 = NoSuchSessionValueError
		actualPresent1   error
		expectedPresent2 = NoSuchSessionValueError
		actualPresent2   error
		expectedPresent3 = error(nil)
		actualPresent3   error
		expectedPresent4 = NoSuchSessionValueError
		actualPresent4   error
		saveValue1       = "saved 1"
		saveValue2       = "saved 2"
		expectedValue3   = saveValue2
		actualValue3     interface{}
	)

	/* run */
	handler1 := NewMinSessionHandler("testHandler1", "/", "example.com")
	handler2 := NewMinSessionHandler("testHandler2", "/", "example.com")

	sesh1, err := handler1.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(localCookies1)
	for _, cookie := range stc1 {
		var badCookie http.Cookie
		localCookies1 = append(localCookies1, cookie)
		badCookie.Name = cookie.Name
		badCookie.Value = cookie.Value + " bar"
		localCookies2 = append(localCookies2, &badCookie)
	}
	_, actualPresent1 = sesh1.GetValue(seshKey)
	err = sesh1.SetValue(seshKey, saveValue1)
	assert.NoError(t, err)

	sesh2, err := handler2.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(localCookies1)
	for _, cookie := range stc2 {
		localCookies1 = append(localCookies1, cookie)
		localCookies2 = append(localCookies2, cookie)
	}
	_, actualPresent2 = sesh2.GetValue(seshKey)
	err = sesh2.SetValue(seshKey, saveValue2)
	assert.NoError(t, err)

	sesh3, err := handler2.GetSession()
	assert.NoError(t, err)
	fwd3, stc3, err3 := sesh3.CookieMask(localCookies2)
	actualValue3, actualPresent3 = sesh3.GetValue(seshKey)

	sesh4, err := handler1.GetSession()
	assert.NoError(t, err)
	fwd4, stc4, err4 := sesh4.CookieMask(localCookies2)
	_, actualPresent4 = sesh4.GetValue(seshKey)

	/* check */
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NoError(t, err4)
	assert.Nil(t, fwd1)
	assert.NotNil(t, fwd2)
	if fwd2 != nil {
		assert.Equal(t, 1, len(fwd2))
	}
	assert.NotNil(t, fwd3)
	if fwd3 != nil {
		assert.Equal(t, 1, len(fwd3))
	}
	assert.NotNil(t, fwd4)
	if fwd4 != nil {
		assert.Equal(t, 1, len(fwd4))
	}
	assert.NotNil(t, stc1)
	if stc1 != nil {
		assert.Equal(t, 1, len(stc1))
	}
	assert.NotNil(t, stc2)
	if stc2 != nil {
		assert.Equal(t, 1, len(stc2))
	}
	assert.Nil(t, stc3)
	assert.NotNil(t, stc4)
	if stc4 != nil {
		assert.Equal(t, 2, len(stc4))
	}
	assert.Equal(t, expectedPresent1, actualPresent1)
	assert.Equal(t, expectedPresent2, actualPresent2)
	assert.Equal(t, expectedPresent3, actualPresent3)
	assert.Equal(t, expectedPresent4, actualPresent4)
	assert.Equal(t, expectedValue3, actualValue3)
}

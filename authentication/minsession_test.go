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
	var local_cookies []*http.Cookie

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	for _, cookie := range stc1 {
		local_cookies = append(local_cookies, cookie)
	}
	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(local_cookies)

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
	var local_cookies []*http.Cookie
	expected_data := make(map[string]interface{})
	expected_key := "test key"

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	_, present1 := sesh1.GetValue(expected_key)
	for _, cookie := range stc1 {
		local_cookies = append(local_cookies, cookie)

		/* intermediate check */
		_, present1 := sesh1.GetValue(expected_key)
		assert.Error(t, present1)
		assert.Equal(t, present1, NoSuchSessionValueError)

		expected_string := "saving data into " + cookie.Name + " cookie"
		expected_data[expected_key] = expected_string
		err := sesh1.SetValue(expected_key, expected_string)
		assert.NoError(t, err)
	}
	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(local_cookies)

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

	actual_data2, present2 := sesh2.GetValue(expected_key)
	assert.NoError(t, present2)
	assert.NotEqual(t, present2, NoSuchSessionValueError)
	assert.Equal(t, expected_data[expected_key], actual_data2)
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
	var local_cookies []*http.Cookie
	var bad_cookie http.Cookie

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	for _, cookie := range stc1 {
		cookie.Value = cookie.Value + "bad"
		bad_cookie.Name = cookie.Name
		bad_cookie.Value = cookie.Value
		local_cookies = append(local_cookies, cookie)
	}
	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(local_cookies)

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
	bad_cookie_deleted := false
	for _, cookie := range stc2 {
		if cookie.Name == bad_cookie.Name {
			assert.Equal(t, bad_cookie.Value, cookie.Value)
			assert.Equal(t, -1, cookie.MaxAge)
			bad_cookie_deleted = true
		}
	}
	assert.True(t, bad_cookie_deleted)
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
	var invalid_cookie http.Cookie
	var local_cookies []*http.Cookie
	invalid_cookie.Name = "testHandler-"
	for i := 0; i < minSessionCookieNameRandSize; i++ {
		invalid_cookie.Name += "ff"
	}
	invalid_cookie.Value = "foo"
	local_cookies = append(local_cookies, &invalid_cookie)

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")
	sesh, err := handler.GetSession()
	assert.NoError(t, err)
	fwd, stc, err := sesh.CookieMask(local_cookies)

	/* check */
	assert.NoError(t, err)
	assert.Nil(t, fwd)
	assert.NotNil(t, stc)
	if stc != nil {
		assert.Equal(t, 2, len(stc))
	}
	bad_cookie_deleted := false
	for _, cookie := range stc {
		if cookie.Name == invalid_cookie.Name {
			assert.Equal(t, invalid_cookie.Value, cookie.Value)
			assert.Equal(t, -1, cookie.MaxAge)
			bad_cookie_deleted = true
		}
	}
	assert.True(t, bad_cookie_deleted)
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
		local_cookies1    []*http.Cookie
		local_cookies2    []*http.Cookie
		sesh_key           = "test key"
		expected_present1 = NoSuchSessionValueError
		actual_present1   error
		expected_present2 = NoSuchSessionValueError
		actual_present2   error
		expected_present3 = error(nil)
		actual_present3   error
		expected_present4 = error(nil)
		actual_present4   error
		expected_present5 = error(nil)
		actual_present5   error
		expected_value3   = "test 3"
		actual_value3     interface{}
		expected_value4   = "test 4"
		actual_value4     interface{}
		expected_value5   = "test 5"
		actual_value5     interface{}
		save_value1       = expected_value4
		save_value2       = expected_value3
		save_value3       = expected_value5
	)

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")

	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(local_cookies1)
	for _, cookie := range stc1 {
		local_cookies1 = append(local_cookies1, cookie)
	}
	_, actual_present1 = sesh1.GetValue(sesh_key)
	err = sesh1.SetValue(sesh_key, save_value1)
	assert.NoError(t, err)

	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(local_cookies2)
	for _, cookie := range stc2 {
		local_cookies2 = append(local_cookies2, cookie)
	}
	_, actual_present2 = sesh2.GetValue(sesh_key)
	err = sesh2.SetValue(sesh_key, save_value2)
	assert.NoError(t, err)

	sesh3, err := handler.GetSession()
	assert.NoError(t, err)
	fwd3, stc3, err3 := sesh3.CookieMask(local_cookies2)
	actual_value3, actual_present3 = sesh3.GetValue(sesh_key)
	err = sesh3.SetValue(sesh_key, save_value3)
	assert.NoError(t, err)

	sesh4, err := handler.GetSession()
	assert.NoError(t, err)
	fwd4, stc4, err4 := sesh4.CookieMask(local_cookies1)
	actual_value4, actual_present4 = sesh4.GetValue(sesh_key)

	sesh5, err := handler.GetSession()
	assert.NoError(t, err)
	fwd5, stc5, err5 := sesh5.CookieMask(local_cookies2)
	actual_value5, actual_present5 = sesh5.GetValue(sesh_key)

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
	assert.Equal(t, expected_present1, actual_present1)
	assert.Equal(t, expected_present2, actual_present2)
	assert.Equal(t, expected_present3, actual_present3)
	assert.Equal(t, expected_present4, actual_present4)
	assert.Equal(t, expected_present5, actual_present5)
	assert.Equal(t, expected_value3, actual_value3)
	assert.Equal(t, expected_value4, actual_value4)
	assert.Equal(t, expected_value5, actual_value5)
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
		good_cookies             []*http.Cookie
		bad_cookies              []*http.Cookie
		bad_cookie               http.Cookie
		sesh_key                  = "test key"
		expected_sesh_present1    = NoSuchSessionValueError
		actual_sesh_present1      error
		expected_sesh_present2    = NoSuchSessionValueError
		actual_sesh_present2      error
		expected_sesh_present3    = NoSuchSessionValueError
		actual_sesh_present3      error
		expected_cookie_present2 = true
		actual_cookie_present2   bool
		expected_cookie_present3 = true
		actual_cookie_present3   bool
		save_value               = "foo"
	)

	/* run */
	handler := NewMinSessionHandler("testHandler", "/", "example.com")

	sesh1, err := handler.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(nil)
	for _, good_cookie := range stc1 {
		good_cookies = append(good_cookies, good_cookie)
		bad_cookie.Name = good_cookie.Name
		bad_cookie.Value = good_cookie.Value + " bar"
		bad_cookies = append(bad_cookies, &bad_cookie)
	}
	_, actual_sesh_present1 = sesh1.GetValue(sesh_key)
	err = sesh1.SetValue(sesh_key, save_value)
	assert.NoError(t, err)

	sesh2, err := handler.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(bad_cookies)
	actual_cookie_present2 = false
	for _, cookie := range stc2 {
		if cookie.Name == bad_cookie.Name {
			actual_cookie_present2 = true
		}
	}
	_, actual_sesh_present2 = sesh2.GetValue(sesh_key)

	sesh3, err := handler.GetSession()
	assert.NoError(t, err)
	fwd3, stc3, err3 := sesh3.CookieMask(good_cookies)
	actual_cookie_present3 = false
	for _, cookie := range stc3 {
		if cookie.Name == bad_cookie.Name {
			actual_cookie_present3 = true
		}
	}
	_, actual_sesh_present3 = sesh3.GetValue(sesh_key)

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
	assert.Equal(t, expected_sesh_present1, actual_sesh_present1)
	assert.Equal(t, expected_sesh_present2, actual_sesh_present2)
	assert.Equal(t, expected_sesh_present3, actual_sesh_present3)
	assert.Equal(t, expected_cookie_present2, actual_cookie_present2)
	assert.Equal(t, expected_cookie_present3, actual_cookie_present3)
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
		local_cookies1    []*http.Cookie
		local_cookies2    []*http.Cookie
		sesh_key           = "test key"
		expected_present1 = NoSuchSessionValueError
		actual_present1   error
		expected_present2 = NoSuchSessionValueError
		actual_present2   error
		expected_present3 = error(nil)
		actual_present3   error
		expected_present4 = NoSuchSessionValueError
		actual_present4   error
		save_value1       = "saved 1"
		save_value2       = "saved 2"
		expected_value3   = save_value2
		actual_value3     interface{}
	)

	/* run */
	handler1 := NewMinSessionHandler("testHandler1", "/", "example.com")
	handler2 := NewMinSessionHandler("testHandler2", "/", "example.com")

	sesh1, err := handler1.GetSession()
	assert.NoError(t, err)
	fwd1, stc1, err1 := sesh1.CookieMask(local_cookies1)
	for _, cookie := range stc1 {
		var bad_cookie http.Cookie
		local_cookies1 = append(local_cookies1, cookie)
		bad_cookie.Name = cookie.Name
		bad_cookie.Value = cookie.Value + " bar"
		local_cookies2 = append(local_cookies2, &bad_cookie)
	}
	_, actual_present1 = sesh1.GetValue(sesh_key)
	err = sesh1.SetValue(sesh_key, save_value1)
	assert.NoError(t, err)

	sesh2, err := handler2.GetSession()
	assert.NoError(t, err)
	fwd2, stc2, err2 := sesh2.CookieMask(local_cookies1)
	for _, cookie := range stc2 {
		local_cookies1 = append(local_cookies1, cookie)
		local_cookies2 = append(local_cookies2, cookie)
	}
	_, actual_present2 = sesh2.GetValue(sesh_key)
	err = sesh2.SetValue(sesh_key, save_value2)
	assert.NoError(t, err)

	sesh3, err := handler2.GetSession()
	assert.NoError(t, err)
	fwd3, stc3, err3 := sesh3.CookieMask(local_cookies2)
	actual_value3, actual_present3 = sesh3.GetValue(sesh_key)

	sesh4, err := handler1.GetSession()
	assert.NoError(t, err)
	fwd4, stc4, err4 := sesh4.CookieMask(local_cookies2)
	_, actual_present4 = sesh4.GetValue(sesh_key)

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
	assert.Equal(t, expected_present1, actual_present1)
	assert.Equal(t, expected_present2, actual_present2)
	assert.Equal(t, expected_present3, actual_present3)
	assert.Equal(t, expected_present4, actual_present4)
	assert.Equal(t, expected_value3, actual_value3)
}

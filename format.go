// Copyright 2016 Liam Stanley <me@liamstanley.io>. All rights reserved.
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package girc

import (
	"bytes"
	"strings"
)

type color struct {
	aliases []string
	val     string
}

var colors = []*color{
	{aliases: []string{"white"}, val: "\x0300"},
	{aliases: []string{"black"}, val: "\x0301"},
	{aliases: []string{"blue", "navy"}, val: "\x0302"},
	{aliases: []string{"green"}, val: "\x0303"},
	{aliases: []string{"red"}, val: "\x0304"},
	{aliases: []string{"brown", "maroon"}, val: "\x0305"},
	{aliases: []string{"purple"}, val: "\x0306"},
	{aliases: []string{"orange", "olive", "gold"}, val: "\x0307"},
	{aliases: []string{"yellow"}, val: "\x0308"},
	{aliases: []string{"lightgreen", "lime"}, val: "\x0309"},
	{aliases: []string{"teal"}, val: "\x0310"},
	{aliases: []string{"cyan"}, val: "\x0311"},
	{aliases: []string{"lightblue", "royal"}, val: "\x0312"},
	{aliases: []string{"lightpurple", "pink", "fuchsia"}, val: "\x0313"},
	{aliases: []string{"grey", "gray"}, val: "\x0314"},
	{aliases: []string{"lightgrey", "silver"}, val: "\x0315"},
	{aliases: []string{"bold", "b"}, val: "\x02"},
	{aliases: []string{"italic", "i"}, val: "\x1d"},
	{aliases: []string{"reset", "r"}, val: "\x0f"},
	{aliases: []string{"clear", "c"}, val: "\x03"},
	{aliases: []string{"reverse"}, val: "\x16"},
	{aliases: []string{"underline", "ul"}, val: "\x1f"},
}

// Format takes color strings like "{red}" and turns them into the resulting
// ASCII color code for IRC.
func Format(text string) string {
	for i := 0; i < len(colors); i++ {
		for a := 0; a < len(colors[i].aliases); a++ {
			text = strings.Replace(text, "{"+colors[i].aliases[a]+"}", colors[i].val, -1)
		}

		// makes parsing small strings slightly slower, but helps longer
		// strings.
		var more bool
		for c := 0; c < len(text); c++ {
			if text[c] == 0x7B {
				more = true
				break
			}
		}
		if !more {
			return text
		}
	}

	return text
}

// StripFormat strips all "{color}" formatting strings from the input text.
// See Format() for more information.
func StripFormat(text string) string {
	for i := 0; i < len(colors); i++ {
		for a := 0; a < len(colors[i].aliases); a++ {
			text = strings.Replace(text, "{"+colors[i].aliases[a]+"}", "", -1)
		}

		// makes parsing small strings slightly slower, but helps longer
		// strings.
		var more bool
		for c := 0; c < len(text); c++ {
			if text[c] == 0x7B {
				more = true
				break
			}
		}
		if !more {
			return text
		}
	}

	return text
}

// StripColors tries to strip all ASCII color codes that are used for IRC.
func StripColors(text string) string {
	for i := 0; i < len(colors); i++ {
		text = strings.Replace(text, colors[i].val, "", -1)
	}

	return text
}

// IsValidChannel validates if channel is an RFC complaint channel or not.
//
// NOTE: If you do not need to validate against servers that support Unicode,
// you may want to ensure that all channel chars are within the range of
// all ASCII printable chars. This function will NOT do that for
// compatibility reasons.
//
//   channel    =  ( "#" / "+" / ( "!" channelid ) / "&" ) chanstring
//                 [ ":" chanstring ]
//   chanstring =  0x01-0x07 / 0x08-0x09 / 0x0B-0x0C / 0x0E-0x1F / 0x21-0x2B
//   chanstring =  / 0x2D-0x39 / 0x3B-0xFF
//                   ; any octet except NUL, BELL, CR, LF, " ", "," and ":"
//   channelid  = 5( 0x41-0x5A / digit )   ; 5( A-Z / 0-9 )
func IsValidChannel(channel string) bool {
	if len(channel) <= 1 || len(channel) > 50 {
		return false
	}

	// #, +, !<channelid>, or &
	// Including "*" in the prefix list, as this is commonly used (e.g. ZNC)
	if bytes.IndexByte([]byte{0x21, 0x23, 0x26, 0x2A, 0x2B}, channel[0]) == -1 {
		return false
	}

	// !<channelid> -- not very commonly supported, but we'll check it anyway.
	// The ID must be 5 chars. This means min-channel size should be:
	//   1 (prefix) + 5 (id) + 1 (+, channel name)
	if channel[0] == 0x21 {
		if len(channel) < 7 {
			return false
		}

		// check for valid ID
		for i := 1; i < 6; i++ {
			if (channel[i] < 0x30 || channel[i] > 0x39) && (channel[i] < 0x41 || channel[i] > 0x5A) {
				return false
			}
		}
	}

	// Check for invalid octets here.
	bad := []byte{0x00, 0x07, 0x0D, 0x0A, 0x20, 0x2C, 0x3A}
	for i := 1; i < len(channel); i++ {
		if bytes.IndexByte(bad, channel[i]) != -1 {
			return false
		}
	}

	return true
}

// IsValidNick validates an IRC nickame. Note that this does not validate
// IRC nickname length.
//
//   nickname =  ( letter / special ) *8( letter / digit / special / "-" )
//   letter   =  0x41-0x5A / 0x61-0x7A
//   digit    =  0x30-0x39
//   special  =  0x5B-0x60 / 0x7B-0x7D
func IsValidNick(nick string) bool {
	if len(nick) <= 0 {
		return false
	}

	// Check the first index. Some characters aren't allowed for the first
	// index of an IRC nickname.
	if nick[0] < 0x41 || nick[0] > 0x7D {
		// a-z, A-Z, and _\[]{}^|
		return false
	}

	for i := 1; i < len(nick); i++ {
		if (nick[i] < 0x41 || nick[i] > 0x7D) && (nick[i] < 0x30 || nick[i] > 0x39) && nick[i] != 0x2D {
			// a-z, A-Z, 0-9, -, and _\[]{}^|
			return false
		}
	}

	return true
}

// IsValidUser validates an IRC ident/username. Note that this does not
// validate IRC ident length.
//
// The validation checks are much like what characters are allowed with an
// IRC nickname (see IsValidNick()), however an ident/username can:
//
// 1. Must either start with alphanumberic char, or "~" then
//    alphanumberic char.
// 2. Contain a "." (period), for use with "first.last".
//    Though, this may not be supported on all networks. Some limit this
//    to only a single period.
//
// Per RFC:
//    user =  1*( %x01-09 / %x0B-0C / %x0E-1F / %x21-3F / %x41-FF )
//            ; any octet except NUL, CR, LF, " " and "@"
func IsValidUser(name string) bool {
	if len(name) <= 0 {
		return false
	}

	// "~" is prepended (commonly) if there was no ident server response.
	if name[0] == 0x7E {
		// Means name only contained "~".
		if len(name) < 2 {
			return false
		}

		name = name[1:]
	}

	// Check to see if the first index is alphanumeric.
	if (name[0] < 0x41 || name[0] > 0x4A) && (name[0] < 0x61 || name[0] > 0x7A) && (name[0] < 0x30 || name[0] > 0x39) {
		return false
	}

	for i := 1; i < len(name); i++ {
		if (name[i] < 0x41 || name[i] > 0x7D) && (name[i] < 0x30 || name[i] > 0x39) && name[i] != 0x2D && name[i] != 0x2E {
			// a-z, A-Z, 0-9, -, and _\[]{}^|
			return false
		}
	}

	return true
}
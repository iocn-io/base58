package base58

func FastEncode(bin []byte) []byte {
	zero := uint8(alphabetIdx0)

	binsz := len(bin)
	var i, j, zcount, high int
	var carry uint32

	for zcount < binsz && bin[zcount] == 0 {
		zcount++
	}

	size := (binsz-zcount)*138/100 + 1
	var buf = make([]byte, size)

	high = size - 1
	for i = zcount; i < binsz; i++ {
		j = size - 1
		for carry = uint32(bin[i]); j > high || carry != 0; j-- {
			carry = carry + 256*uint32(buf[j])
			buf[j] = byte(carry % 58)
			carry /= 58
		}
		high = j
	}

	for j = 0; j < size && buf[j] == 0; j++ {
	}

	var code = make([]byte, size-j+zcount)

	if zcount != 0 {
		for i = 0; i < zcount; i++ {
			code[i] = zero
		}
	}

	for i = zcount; j < size; i++ {
		code[i] = alphabet[buf[j]]
		j++
	}

	return code
}

func FastDecode(bytes []byte) []byte {
	if len(bytes) == 0 {
		return nil //, fmt.Errorf("zero length string")
	}

	var (
		t        uint64
		zmask, c uint32
		zcount   int

		b58sz = len(bytes)

		outisz    = (b58sz + 3) / 4 // check to see if we need to change this buffer size to optimize
		binu      = make([]byte, (b58sz+3)*3)
		bytesleft = b58sz % 4

		zero = byte(alphabetIdx0)
	)

	if bytesleft > 0 {
		zmask = (0xffffffff << uint32(bytesleft*8))
	} else {
		bytesleft = 4
	}

	var outi = make([]uint32, outisz)

	for i := 0; i < b58sz && bytes[i] == zero; i++ {
		zcount++
	}

	for _, r := range bytes {
		if r > 255 {
			return nil //, fmt.Errorf("High-bit set on invalid digit")
		}
		if b58[r] == 255 {
			return nil //, fmt.Errorf("Invalid base58 digit (%q)", r)
		}

		c = uint32(b58[r])

		for j := (outisz - 1); j >= 0; j-- {
			t = uint64(outi[j])*58 + uint64(c)
			c = uint32(t>>32) & 0x3f
			outi[j] = uint32(t & 0xffffffff)
		}

		if c > 0 {
			return nil //, fmt.Errorf("Output number too big (carry to the next int32)")
		}

		if outi[0]&zmask != 0 {
			return nil //, fmt.Errorf("Output number too big (last int32 filled too far)")
		}
	}

	var j, cnt int
	for j, cnt = 0, 0; j < outisz; j++ {
		for mask := byte(bytesleft-1) * 8; mask <= 0x18; mask, cnt = mask-8, cnt+1 {
			binu[cnt] = byte(outi[j] >> mask)
		}
		if j == 0 {
			bytesleft = 4 // because it could be less than 4 the first time through
		}
	}

	for n, v := range binu {
		if v > 0 {
			start := n - zcount
			if start < 0 {
				start = 0
			}
			return binu[start:cnt]
		}
	}
	return binu[:cnt]
}

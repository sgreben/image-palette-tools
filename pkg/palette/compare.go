package palette

import "image/color"

func LessLSH(cs []color.RGBA, i, j int) bool {
	hi, si, li := hsl(cs[i].R, cs[i].G, cs[i].B)
	hj, sj, lj := hsl(cs[j].R, cs[j].G, cs[j].B)
	if li == lj {
		if si == sj {
			return hi < hj
		}
		return si < sj
	}
	return li < lj
}

func LessHLS(cs []color.RGBA, i, j int) bool {
	hi, si, li := hsl(cs[i].R, cs[i].G, cs[i].B)
	hj, sj, lj := hsl(cs[j].R, cs[j].G, cs[j].B)
	if hi == hj {
		if li == lj {
			return si < sj
		}
		return li < lj
	}
	return hi < hj
}

func LessLHS(cs []color.RGBA, i, j int) bool {
	hi, si, li := hsl(cs[i].R, cs[i].G, cs[i].B)
	hj, sj, lj := hsl(cs[j].R, cs[j].G, cs[j].B)
	if li == lj {
		if hi == hj {
			return si < sj
		}
		return hi < hj
	}
	return li < lj
}

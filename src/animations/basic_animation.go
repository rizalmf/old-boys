package animations

type Animation struct {
	First        int
	Last         int
	Step         int     // how many indics do we move per frame
	SpeedInTps   float32 // how many ticks before next frame
	FrameCounter float32
	frame        int
	y            int // horizontal spritesheet purpose
}

func NewAnimationVertical(first, last, step int, speedInTps float32) *Animation {
	return &Animation{
		First:        first,
		Last:         last,
		Step:         step,
		SpeedInTps:   speedInTps,
		FrameCounter: speedInTps, // initialize frame counter to speedInTps
		frame:        first,
	}
}

func NewAnimationHorizotal(y, last int, speedInTps float32) *Animation {
	return &Animation{
		First:        0,
		Last:         last,
		Step:         1,
		SpeedInTps:   speedInTps,
		FrameCounter: speedInTps, // initialize frame counter to speedInTps
		frame:        0,
		y:            y,
	}
}

func (a *Animation) Frame() (int, int) {
	return a.frame, a.y
}

func (a *Animation) Update() {
	a.FrameCounter -= 1
	if a.FrameCounter <= 0 {
		a.FrameCounter = a.SpeedInTps
		a.frame += a.Step

		if a.frame > a.Last {
			// loop back to the first frame
			a.frame = a.First
		}
	}

}

package algorithms

type WeightedAStar struct{}

func (wa *WeightedAStar) Run(input AlgorithmInput) AlgorithmOutput {
	w := input.Params.W
	if w < 1.0 {
		w = 1.0
	}
	return runAStarFamily(input, w)
}

package pets

import "errors"

type PetInterface interface {
	SetSpecies(species string) *Pet
	SetBreed(breed string) *Pet
	SetMinWeight(minWeight int) *Pet
	SetMaxWeight(maxWeight int) *Pet
	SetAverageWeight(averageWeight int) *Pet
	SetWeight(weight int) *Pet
	SetDescription(description string) *Pet
	SetLifeSpan(lifeSpan int) *Pet
	SetGeographicOrigin(geographicOrigin string) *Pet
	SetColor(color string) *Pet
	SetAge(age int) *Pet
	SetAgeEstimated(ageEstimated int) *Pet
}

func (p *Pet) SetSpecies(species string) *Pet {
	p.Species = species
	return p
}

func (p *Pet) SetBreed(breed string) *Pet {
	p.Breed = breed
	return p
}

func (p *Pet) SetMinWeight(minWeight int) *Pet {
	p.MinWeight = minWeight
	return p
}

func (p *Pet) SetMaxWeight(maxWeight int) *Pet {
	p.MaxWeight = maxWeight
	return p
}

func (p *Pet) SetAverageWeight(averageWeight int) *Pet {
	p.AverageWeight = averageWeight
	return p
}

func (p *Pet) SetWeight(weight int) *Pet {
	p.Weight = weight
	return p
}

func (p *Pet) SetDescription(description string) *Pet {
	p.Description = description
	return p
}

func (p *Pet) SetLifeSpan(lifeSpan int) *Pet {
	p.LifeSpan = lifeSpan
	return p
}

func (p *Pet) SetGeographicOrigin(geographicOrigin string) *Pet {
	p.GeographicOrigin = geographicOrigin
	return p
}

func (p *Pet) SetColor(color string) *Pet {
	p.Color = color
	return p
}

func (p *Pet) SetAge(age int) *Pet {
	p.Age = age
	return p
}

func (p *Pet) SetAgeEstimated(ageEstimated int) *Pet {
	p.AgeEstimated = ageEstimated
	return p
}

func (p *Pet) Build() (*Pet, error) {
	if p.MinWeight > p.MaxWeight {
		return nil, errors.New("minimum weight is greater than maximum weight")
	}

	p.AverageWeight = (p.MinWeight + p.MaxWeight) / 2

	return p, nil
}

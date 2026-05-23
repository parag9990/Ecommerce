package domain

type ValidationOptions struct {
	StrictAttributeSchema         bool
	RequirePrimaryImageForPublish bool
	MaxImagesPerProduct           int
	MaxVariantsPerProduct         int
	DefaultCurrency               string
}

func DefaultValidationOptions() ValidationOptions {
	return ValidationOptions{
		StrictAttributeSchema:         true,
		RequirePrimaryImageForPublish: false,
		MaxImagesPerProduct:           50,
		MaxVariantsPerProduct:         250,
		DefaultCurrency:               "INR",
	}
}

func (o ValidationOptions) normalized() ValidationOptions {
	defaults := DefaultValidationOptions()
	if o.MaxImagesPerProduct <= 0 {
		o.MaxImagesPerProduct = defaults.MaxImagesPerProduct
	}
	if o.MaxVariantsPerProduct <= 0 {
		o.MaxVariantsPerProduct = defaults.MaxVariantsPerProduct
	}
	if o.DefaultCurrency == "" {
		o.DefaultCurrency = defaults.DefaultCurrency
	}
	return o
}

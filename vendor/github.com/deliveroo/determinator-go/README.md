# Determinator

This is a port of the [Determinator ruby gem](https://github.com/deliveroo/determinator) — used for categorizing "actors" (most usually users) in your application for running A/B tests and slowly rolling out new features.

#### Getting help

For Deliveroo Employees:

- Many people contribute to Determinator and Florence. We hang out in [this Slack channel](https://deliveroo.slack.com/app_redirect?channel=florence_wg)
- [This JIRA board](https://deliveroo.atlassian.net/secure/RapidBoard.jspa?rapidView=156) covers pieces of work that are planned or in-flight
- [This Workplace group](https://deliveroo.facebook.com/groups/1893254264328414/) holds more general discussions about the Florence ecosystem

At the moment we can only promise support for Determinator within Deliveroo, but if you add [issues to this github repo](https://github.com/deliveroo/determinator/issues) we'll try and help if we can!

## Basic Use

Determinator-go can be used to determine whether a **feature flag** or **experiment** is on or off for the current actor (or user) and, for experiments, which **variant** they should see.

Determinator-go is made of up of two interfaces Retriever and Feature.

Retriever is used to retrieve a feature from the file system, over http, from a cache etc. For example to retrieve a feature
from the file system you would use.

```golang

rootPath := "/path_to/feature_folder"
fileRetriever := determinator.NewFileRetriever(rootPath)

featureName := "feature"

feature, err := fileRetriever.Retrieve(featureName)
```

Feature is used to determine whether a feature is a experiment and whether a actor should have access to the feature.

```golang
// checks whether a feature is switched on for an actor
feature.IsFeatureFlagOn(actor)


// for experiments
// checks whether a feature is part of an experiment
featue.IsExperiment()

//checks which variant of an experiment a actor belongs to
switch feature.WhichVariant(actor) {
	case "A":
		fmt.Println("%s is in A!", actor.ID)
	case "B":
		fmt.Println("%s is in B!", actor.ID)
	default:
		fmt.Printf("%s is not part of an experiment", actor.ID)
	}
}
```

## Context objects

To be consistent in determinations and avoid mistakes, we recommend using structured buckets with context objects.

To set up a context object, e.g. at the start of a request:

```golang
determinatorContext := determinator.Context{
	Actors: determinator.Actors{
		Request: &models.Request{
			UID:      "123",
			City:     models.City{ID: 1, UName: "london"},
			Country:  models.Country{ID: 1, TLD: "uk"},
			Zone:     models.Zone{ID: 1, Code: "test"},
			Customer: &models.RequestCustomer{GUID: "abc"},
			Platform: "web",
		},
		Customer: &models.Customer{
			ID: 321,
		},
	},
	Retriever: retriever
}
```

To help pass it around, you can store it on a `context.Context`:

```golang
ctx = determinator.SetContext(ctx, determinatorContext)
```

And later retrieve it and determinate as needed:

```golang
determinatorContext, ok := determinator.GetContext(ctx)
res, err := determinatorContext.WhichVariant("my-feature")
```

This will set the determination key based on the feature's `StructuredBucket` setting, and pass all the other attributes as params.

## Tracking

The library allows user to track the determinations that have been made as well as a middleware that wraps the functionality
around a http request, allowing logging them at the end of the request (including some useful request metrics).

```golang

featureTracker := CreateFeatureTracker(feature)

featureTracker.Determinate(actor1)
featureTracker.Determinate(actor2)

determinations := featureTracker.GetDeterminations()

```

To track the determinations of the features you obtained from your favourite retriever

```golang

retriver := NewFileRetriever("/my/path")
trackingRetriever := ConstructTrackingRetriever(retriever)

feature1 := trackingRetriever.Retriever("feature1")
feature2 := trackingRetriever.Retriever("feature2")

//do determinations with feature1 and feature2

//get all determinations made by feature1 and feature2
determinations := trackingRetriever.CollectDeterminations()

```

To use it as a middleware eg with the gorilla framework.

```golang

handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dt, ok := determinator.TrackingRetrieverFromContext(r.Context())
		if !ok {
			fmt.Errorf("could not get TrackingRetriever from context")
		}

		ret, err := dt.Retrieve("test-feature")
		if err != nil {
			fmt.Errorf("got error %s trying to retrieve the feature", err)
		}

    //use the retriver
	})

retriver := NewFileRetriever("/my/path")

//create router...

router.Use(determinator.Determinating(&retriever))

router.HandleFunc("/", fakeHandler)

//handle requests...
```

The featureTracker and trackingRetriever are thread safe.

## Generating code from your feature files

To more easily interact with Determinator, [cmd/feature-codegen](cmd/feature-codegen) has been added to generate a handy interface for verifying whether a flag is set.

It can be executed as a one-off as such:

```sh
$ ../determinator-go/cmd/feature-codegen/feature-codegen
```

But it's recommended to be embedded into your build. For instance, in a file `pkg/florence/features.go`:

```go
// if you want to install the `feature-codegen` binary into your $PATH via `go install`
//go:generate feature-codegen -dest .
// or if you want to track your dependencies in your `go.mod`, i.e. https://www.jvt.me/posts/2022/06/15/go-tools-dependency-management/
//go:generate go run github.com/deliveroo/determinator-go/cmd/feature-codegen -dest .
package florence
```

Then run:

```sh
go generate ./...
```

Check out the source of `feature-codegen` for more information about the flags that are available.

## Set up:

After checking out the project, make sure to update submodules

```
git submodule update --init --recursive
```

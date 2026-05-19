package comments

import "github.com/augmentable-dev/lege"

// CStyleCommentOptions ...
var CStyleCommentOptions *lege.ParseOptions = &lege.ParseOptions{
	Boundaries: []lege.Boundary{
		{
			Start: "//",
			End:   "\n",
		},
		{
			Start: "/*",
			End:   "*/",
		},
	},
}

// HashStyleCommentOptions ...
var HashStyleCommentOptions *lege.ParseOptions = &lege.ParseOptions{
	Boundaries: []lege.Boundary{
		{
			Start: "#",
			End:   "\n",
		},
	},
}

// LispStyleCommentOptions ..
var LispStyleCommentOptions *lege.ParseOptions = &lege.ParseOptions{
	Boundaries: []lege.Boundary{
		{
			Start: ";",
			End:   "\n",
		},
	},
}

// Language is a source language (i.e. "Go")
type Language string

// LanguageParseOptions keeps track of source languages and their corresponding comment options
var LanguageParseOptions map[Language]*lege.ParseOptions = map[Language]*lege.ParseOptions{
	"C":            CStyleCommentOptions,
	"C#":           CStyleCommentOptions,
	"C++":          CStyleCommentOptions,
	"Common Lisp":  LispStyleCommentOptions,
	"Emacs Lisp":   LispStyleCommentOptions,
	"Go":           CStyleCommentOptions,
	"Groovy":       CStyleCommentOptions,
	"Haskell":      {Boundaries: []lege.Boundary{{Start: "--", End: "\n"}, {Start: "{-", End: "-}"}}},
	"Java":         CStyleCommentOptions,
	"JavaScript":   CStyleCommentOptions,
	"Objective-C":  CStyleCommentOptions,
	"PHP":          {Boundaries: append(CStyleCommentOptions.Boundaries, HashStyleCommentOptions.Boundaries...)},
	"Python":       HashStyleCommentOptions,
	"R":            HashStyleCommentOptions,
	"Ruby":         HashStyleCommentOptions,
	"Shell":        HashStyleCommentOptions,
	"Swift":        CStyleCommentOptions,
	"TypeScript":   CStyleCommentOptions,
	"Visual Basic": {Boundaries: []lege.Boundary{{Start: "'", End: "\n"}}},
	// Track lege precedence limitations in https://github.com/MTG-Thomas/tickgit/issues/9.
	// so lines beginning with /// or //! will be picked up by this start // and include a / or ! preceding the comment
	"Kotlin": CStyleCommentOptions,
	"Rust":   {Boundaries: []lege.Boundary{{Start: "///", End: "\n"}, {Start: "//!", End: "\n"}, {Start: "//", End: "\n"}}},

	// Track lege boundary limitations in https://github.com/MTG-Thomas/tickgit/issues/9.
	// Multi-line comments are not getting picked up...
	"Elixir": HashStyleCommentOptions,
	"Julia":  {Boundaries: []lege.Boundary{{Start: "#=", End: "=#"}, {Start: "#", End: "\n"}}},
}

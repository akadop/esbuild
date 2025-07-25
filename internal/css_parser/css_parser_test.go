package css_parser

import (
	"fmt"
	"testing"

	"github.com/evanw/esbuild/internal/ast"
	"github.com/evanw/esbuild/internal/compat"
	"github.com/evanw/esbuild/internal/config"
	"github.com/evanw/esbuild/internal/css_printer"
	"github.com/evanw/esbuild/internal/logger"
	"github.com/evanw/esbuild/internal/test"
)

func expectPrintedCommon(t *testing.T, name string, contents string, expected string, expectedLog string, loader config.Loader, options config.Options) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Helper()
		log := logger.NewDeferLog(logger.DeferLogNoVerboseOrDebug, nil)
		tree := Parse(log, test.SourceForTest(contents), OptionsFromConfig(loader, &options))
		msgs := log.Done()
		text := ""
		for _, msg := range msgs {
			text += msg.String(logger.OutputOptions{}, logger.TerminalInfo{})
		}
		test.AssertEqualWithDiff(t, text, expectedLog)
		symbols := ast.NewSymbolMap(1)
		symbols.SymbolsForSource[0] = tree.Symbols
		result := css_printer.Print(tree, symbols, css_printer.Options{
			MinifyWhitespace: options.MinifyWhitespace,
		})
		test.AssertEqualWithDiff(t, string(result.CSS), expected)
	})
}

func expectPrinted(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents, contents, expected, expectedLog, config.LoaderCSS, config.Options{})
}

func expectPrintedLocal(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [local]", contents, expected, expectedLog, config.LoaderLocalCSS, config.Options{})
}

func expectPrintedLower(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [lower]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		UnsupportedCSSFeatures: ^compat.CSSFeature(0),
	})
}

func expectPrintedLowerUnsupported(t *testing.T, unsupportedCSSFeatures compat.CSSFeature, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [lower]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		UnsupportedCSSFeatures: unsupportedCSSFeatures,
	})
}

func expectPrintedWithAllPrefixes(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [prefixed]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		CSSPrefixData: compat.CSSPrefixData(map[compat.Engine]compat.Semver{
			compat.Chrome:  {Parts: []int{0}},
			compat.Edge:    {Parts: []int{0}},
			compat.Firefox: {Parts: []int{0}},
			compat.IE:      {Parts: []int{0}},
			compat.IOS:     {Parts: []int{0}},
			compat.Opera:   {Parts: []int{0}},
			compat.Safari:  {Parts: []int{0}},
		}),
	})
}

func expectPrintedMinify(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [minify]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		MinifyWhitespace: true,
	})
}

func expectPrintedMangle(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [mangle]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		MinifySyntax: true,
	})
}

func expectPrintedLowerMangle(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [lower, mangle]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		UnsupportedCSSFeatures: ^compat.CSSFeature(0),
		MinifySyntax:           true,
	})
}

func expectPrintedMangleMinify(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [mangle, minify]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		MinifySyntax:     true,
		MinifyWhitespace: true,
	})
}

func expectPrintedLowerMinify(t *testing.T, contents string, expected string, expectedLog string) {
	t.Helper()
	expectPrintedCommon(t, contents+" [lower, minify]", contents, expected, expectedLog, config.LoaderCSS, config.Options{
		UnsupportedCSSFeatures: ^compat.CSSFeature(0),
		MinifyWhitespace:       true,
	})
}

func TestSingleLineComment(t *testing.T) {
	expectPrinted(t, "a, // a\nb // b\n{}", "a, // a b // b {\n}\n",
		"<stdin>: WARNING: Comments in CSS use \"/* ... */\" instead of \"//\"\n"+
			"<stdin>: WARNING: Comments in CSS use \"/* ... */\" instead of \"//\"\n")
	expectPrinted(t, "a, ///// a /////\n{}", "a, ///// a ///// {\n}\n",
		"<stdin>: WARNING: Comments in CSS use \"/* ... */\" instead of \"//\"\n")
}

func TestEscapes(t *testing.T) {
	// TIdent
	expectPrinted(t, "a { value: id\\65nt }", "a {\n  value: ident;\n}\n", "")
	expectPrinted(t, "a { value: \\69 dent }", "a {\n  value: ident;\n}\n", "")
	expectPrinted(t, "a { value: \\69dent }", "a {\n  value: \u69DEnt;\n}\n", "")
	expectPrinted(t, "a { value: \\2cx }", "a {\n  value: \\,x;\n}\n", "")
	expectPrinted(t, "a { value: \\,x }", "a {\n  value: \\,x;\n}\n", "")
	expectPrinted(t, "a { value: x\\2c }", "a {\n  value: x\\,;\n}\n", "")
	expectPrinted(t, "a { value: x\\, }", "a {\n  value: x\\,;\n}\n", "")
	expectPrinted(t, "a { value: x\\0 }", "a {\n  value: x\uFFFD;\n}\n", "")
	expectPrinted(t, "a { value: x\\1 }", "a {\n  value: x\\\x01;\n}\n", "")
	expectPrinted(t, "a { value: x\x00 }", "a {\n  value: x\uFFFD;\n}\n", "")
	expectPrinted(t, "a { value: x\x01 }", "a {\n  value: x\x01;\n}\n", "")

	// THash
	expectPrinted(t, "a { value: #0h\\61sh }", "a {\n  value: #0hash;\n}\n", "")
	expectPrinted(t, "a { value: #\\30hash }", "a {\n  value: #0hash;\n}\n", "")
	expectPrinted(t, "a { value: #\\2cx }", "a {\n  value: #\\,x;\n}\n", "")
	expectPrinted(t, "a { value: #\\,x }", "a {\n  value: #\\,x;\n}\n", "")

	// THashID
	expectPrinted(t, "a { value: #h\\61sh }", "a {\n  value: #hash;\n}\n", "")
	expectPrinted(t, "a { value: #\\68 ash }", "a {\n  value: #hash;\n}\n", "")
	expectPrinted(t, "a { value: #\\68ash }", "a {\n  value: #\u068Ash;\n}\n", "")
	expectPrinted(t, "a { value: #x\\2c }", "a {\n  value: #x\\,;\n}\n", "")
	expectPrinted(t, "a { value: #x\\, }", "a {\n  value: #x\\,;\n}\n", "")

	// TFunction
	expectPrinted(t, "a { value: f\\6e() }", "a {\n  value: fn();\n}\n", "")
	expectPrinted(t, "a { value: \\66n() }", "a {\n  value: fn();\n}\n", "")
	expectPrinted(t, "a { value: \\2cx() }", "a {\n  value: \\,x();\n}\n", "")
	expectPrinted(t, "a { value: \\,x() }", "a {\n  value: \\,x();\n}\n", "")
	expectPrinted(t, "a { value: x\\2c() }", "a {\n  value: x\\,();\n}\n", "")
	expectPrinted(t, "a { value: x\\,() }", "a {\n  value: x\\,();\n}\n", "")

	// TString
	expectPrinted(t, "a { value: 'a\\62 c' }", "a {\n  value: \"abc\";\n}\n", "")
	expectPrinted(t, "a { value: 'a\\62c' }", "a {\n  value: \"a\u062C\";\n}\n", "")
	expectPrinted(t, "a { value: '\\61 bc' }", "a {\n  value: \"abc\";\n}\n", "")
	expectPrinted(t, "a { value: '\\61bc' }", "a {\n  value: \"\u61BC\";\n}\n", "")
	expectPrinted(t, "a { value: '\\2c' }", "a {\n  value: \",\";\n}\n", "")
	expectPrinted(t, "a { value: '\\,' }", "a {\n  value: \",\";\n}\n", "")
	expectPrinted(t, "a { value: '\\0' }", "a {\n  value: \"\uFFFD\";\n}\n", "")
	expectPrinted(t, "a { value: '\\1' }", "a {\n  value: \"\x01\";\n}\n", "")
	expectPrinted(t, "a { value: '\x00' }", "a {\n  value: \"\uFFFD\";\n}\n", "")
	expectPrinted(t, "a { value: '\x01' }", "a {\n  value: \"\x01\";\n}\n", "")

	// TURL
	expectPrinted(t, "a { value: url(a\\62 c) }", "a {\n  value: url(abc);\n}\n", "")
	expectPrinted(t, "a { value: url(a\\62c) }", "a {\n  value: url(a\u062C);\n}\n", "")
	expectPrinted(t, "a { value: url(\\61 bc) }", "a {\n  value: url(abc);\n}\n", "")
	expectPrinted(t, "a { value: url(\\61bc) }", "a {\n  value: url(\u61BC);\n}\n", "")
	expectPrinted(t, "a { value: url(\\2c) }", "a {\n  value: url(,);\n}\n", "")
	expectPrinted(t, "a { value: url(\\,) }", "a {\n  value: url(,);\n}\n", "")

	// TAtKeyword
	expectPrinted(t, "a { value: @k\\65yword }", "a {\n  value: @keyword;\n}\n", "")
	expectPrinted(t, "a { value: @\\6b eyword }", "a {\n  value: @keyword;\n}\n", "")
	expectPrinted(t, "a { value: @\\6beyword }", "a {\n  value: @\u06BEyword;\n}\n", "")
	expectPrinted(t, "a { value: @\\2cx }", "a {\n  value: @\\,x;\n}\n", "")
	expectPrinted(t, "a { value: @\\,x }", "a {\n  value: @\\,x;\n}\n", "")
	expectPrinted(t, "a { value: @x\\2c }", "a {\n  value: @x\\,;\n}\n", "")
	expectPrinted(t, "a { value: @x\\, }", "a {\n  value: @x\\,;\n}\n", "")

	// TDimension
	expectPrinted(t, "a { value: 10\\65m }", "a {\n  value: 10em;\n}\n", "")
	expectPrinted(t, "a { value: 10p\\32x }", "a {\n  value: 10p2x;\n}\n", "")
	expectPrinted(t, "a { value: 10e\\32x }", "a {\n  value: 10\\65 2x;\n}\n", "")
	expectPrinted(t, "a { value: 10e-\\32x }", "a {\n  value: 10\\65-2x;\n}\n", "")
	expectPrinted(t, "a { value: 10E\\32x }", "a {\n  value: 10\\45 2x;\n}\n", "")
	expectPrinted(t, "a { value: 10E-\\32x }", "a {\n  value: 10\\45-2x;\n}\n", "")
	expectPrinted(t, "a { value: 10e1e\\32x }", "a {\n  value: 10e1e2x;\n}\n", "")
	expectPrinted(t, "a { value: 10e1e-\\32x }", "a {\n  value: 10e1e-2x;\n}\n", "")
	expectPrinted(t, "a { value: 10e1E\\32x }", "a {\n  value: 10e1E2x;\n}\n", "")
	expectPrinted(t, "a { value: 10e1E-\\32x }", "a {\n  value: 10e1E-2x;\n}\n", "")
	expectPrinted(t, "a { value: 10E1e\\32x }", "a {\n  value: 10E1e2x;\n}\n", "")
	expectPrinted(t, "a { value: 10E1e-\\32x }", "a {\n  value: 10E1e-2x;\n}\n", "")
	expectPrinted(t, "a { value: 10E1E\\32x }", "a {\n  value: 10E1E2x;\n}\n", "")
	expectPrinted(t, "a { value: 10E1E-\\32x }", "a {\n  value: 10E1E-2x;\n}\n", "")
	expectPrinted(t, "a { value: 10\\32x }", "a {\n  value: 10\\32x;\n}\n", "")
	expectPrinted(t, "a { value: 10\\2cx }", "a {\n  value: 10\\,x;\n}\n", "")
	expectPrinted(t, "a { value: 10\\,x }", "a {\n  value: 10\\,x;\n}\n", "")
	expectPrinted(t, "a { value: 10x\\2c }", "a {\n  value: 10x\\,;\n}\n", "")
	expectPrinted(t, "a { value: 10x\\, }", "a {\n  value: 10x\\,;\n}\n", "")

	// This must remain unescaped. See https://github.com/evanw/esbuild/issues/2677
	expectPrinted(t, "@font-face { unicode-range: U+0e2e-0e2f }", "@font-face {\n  unicode-range: U+0e2e-0e2f;\n}\n", "")

	// RDeclaration
	colorWarning := "<stdin>: WARNING: \",olor\" is not a known CSS property\nNOTE: Did you mean \"color\" instead?\n"
	expectPrintedMangle(t, "a { c\\6flor: #f00 }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { \\63olor: #f00 }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { \\2color: #f00 }", "a {\n  \\,olor: #f00;\n}\n", colorWarning)
	expectPrintedMangle(t, "a { \\,olor: #f00 }", "a {\n  \\,olor: #f00;\n}\n", colorWarning)

	// RUnknownAt
	expectPrinted(t, "@unknown;", "@unknown;\n", "")
	expectPrinted(t, "@u\\6eknown;", "@unknown;\n", "")
	expectPrinted(t, "@\\75nknown;", "@unknown;\n", "")
	expectPrinted(t, "@u\\2cnknown;", "@u\\,nknown;\n", "")
	expectPrinted(t, "@u\\,nknown;", "@u\\,nknown;\n", "")
	expectPrinted(t, "@\\2cunknown;", "@\\,unknown;\n", "")
	expectPrinted(t, "@\\,unknown;", "@\\,unknown;\n", "")

	// RAtKeyframes
	expectPrinted(t, "@k\\65yframes abc { from {} }", "@keyframes abc {\n  from {\n  }\n}\n", "")
	expectPrinted(t, "@keyframes \\61 bc { from {} }", "@keyframes abc {\n  from {\n  }\n}\n", "")
	expectPrinted(t, "@keyframes a\\62 c { from {} }", "@keyframes abc {\n  from {\n  }\n}\n", "")
	expectPrinted(t, "@keyframes abc { \\66rom {} }", "@keyframes abc {\n  from {\n  }\n}\n", "")
	expectPrinted(t, "@keyframes a\\2c c { \\66rom {} }", "@keyframes a\\,c {\n  from {\n  }\n}\n", "")
	expectPrinted(t, "@keyframes a\\,c { \\66rom {} }", "@keyframes a\\,c {\n  from {\n  }\n}\n", "")

	// RAtNamespace
	namespaceWarning := "<stdin>: WARNING: \"@namespace\" rules are not supported\n"
	expectPrinted(t, "@n\\61mespace ns 'path';", "@namespace ns \"path\";\n", namespaceWarning)
	expectPrinted(t, "@namespace \\6es 'path';", "@namespace ns \"path\";\n", namespaceWarning)
	expectPrinted(t, "@namespace ns 'p\\61th';", "@namespace ns \"path\";\n", namespaceWarning)
	expectPrinted(t, "@namespace \\2cs 'p\\61th';", "@namespace \\,s \"path\";\n", namespaceWarning)
	expectPrinted(t, "@namespace \\,s 'p\\61th';", "@namespace \\,s \"path\";\n", namespaceWarning)

	// CompoundSelector
	expectPrinted(t, "* {}", "* {\n}\n", "")
	expectPrinted(t, "*|div {}", "*|div {\n}\n", "")
	expectPrinted(t, "\\2a {}", "\\* {\n}\n", "")
	expectPrinted(t, "\\2a|div {}", "\\*|div {\n}\n", "")
	expectPrinted(t, "\\2d {}", "\\- {\n}\n", "")
	expectPrinted(t, "\\2d- {}", "-- {\n}\n", "")
	expectPrinted(t, "-\\2d {}", "-- {\n}\n", "")
	expectPrinted(t, "\\2d 123 {}", "\\-123 {\n}\n", "")

	// SSHash
	expectPrinted(t, "#h\\61sh {}", "#hash {\n}\n", "")
	expectPrinted(t, "#\\2chash {}", "#\\,hash {\n}\n", "")
	expectPrinted(t, "#\\,hash {}", "#\\,hash {\n}\n", "")
	expectPrinted(t, "#\\2d {}", "#\\- {\n}\n", "")
	expectPrinted(t, "#\\2d- {}", "#-- {\n}\n", "")
	expectPrinted(t, "#-\\2d {}", "#-- {\n}\n", "")
	expectPrinted(t, "#\\2d 123 {}", "#\\-123 {\n}\n", "")
	expectPrinted(t, "#\\61hash {}", "#ahash {\n}\n", "")
	expectPrinted(t, "#\\30hash {}", "#\\30hash {\n}\n", "")
	expectPrinted(t, "#0\\2chash {}", "#0\\,hash {\n}\n", "<stdin>: WARNING: Unexpected \"#0\\\\2chash\"\n")
	expectPrinted(t, "#0\\,hash {}", "#0\\,hash {\n}\n", "<stdin>: WARNING: Unexpected \"#0\\\\,hash\"\n")

	// SSClass
	expectPrinted(t, ".cl\\61ss {}", ".class {\n}\n", "")
	expectPrinted(t, ".\\2c class {}", ".\\,class {\n}\n", "")
	expectPrinted(t, ".\\,class {}", ".\\,class {\n}\n", "")

	// SSPseudoClass
	expectPrinted(t, ":pseudocl\\61ss {}", ":pseudoclass {\n}\n", "")
	expectPrinted(t, ":pseudo\\2c class {}", ":pseudo\\,class {\n}\n", "")
	expectPrinted(t, ":pseudo\\,class {}", ":pseudo\\,class {\n}\n", "")
	expectPrinted(t, ":pseudo(cl\\61ss) {}", ":pseudo(class) {\n}\n", "")
	expectPrinted(t, ":pseudo(cl\\2css) {}", ":pseudo(cl\\,ss) {\n}\n", "")
	expectPrinted(t, ":pseudo(cl\\,ss) {}", ":pseudo(cl\\,ss) {\n}\n", "")

	// SSAttribute
	expectPrinted(t, "[\\61ttr] {}", "[attr] {\n}\n", "")
	expectPrinted(t, "[\\2c attr] {}", "[\\,attr] {\n}\n", "")
	expectPrinted(t, "[\\,attr] {}", "[\\,attr] {\n}\n", "")
	expectPrinted(t, "[attr\\7e=x] {}", "[attr\\~=x] {\n}\n", "")
	expectPrinted(t, "[attr\\~=x] {}", "[attr\\~=x] {\n}\n", "")
	expectPrinted(t, "[attr=\\2c] {}", "[attr=\",\"] {\n}\n", "")
	expectPrinted(t, "[attr=\\,] {}", "[attr=\",\"] {\n}\n", "")
	expectPrinted(t, "[attr=\"-\"] {}", "[attr=\"-\"] {\n}\n", "")
	expectPrinted(t, "[attr=\"--\"] {}", "[attr=--] {\n}\n", "")
	expectPrinted(t, "[attr=\"-a\"] {}", "[attr=-a] {\n}\n", "")
	expectPrinted(t, "[\\6es|attr] {}", "[ns|attr] {\n}\n", "")
	expectPrinted(t, "[ns|\\61ttr] {}", "[ns|attr] {\n}\n", "")
	expectPrinted(t, "[\\2cns|attr] {}", "[\\,ns|attr] {\n}\n", "")
	expectPrinted(t, "[ns|\\2c attr] {}", "[ns|\\,attr] {\n}\n", "")
	expectPrinted(t, "[*|attr] {}", "[*|attr] {\n}\n", "")
	expectPrinted(t, "[\\2a|attr] {}", "[\\*|attr] {\n}\n", "")
}

func TestString(t *testing.T) {
	expectPrinted(t, "a:after { content: 'a\\\rb' }", "a:after {\n  content: \"ab\";\n}\n", "")
	expectPrinted(t, "a:after { content: 'a\\\nb' }", "a:after {\n  content: \"ab\";\n}\n", "")
	expectPrinted(t, "a:after { content: 'a\\\fb' }", "a:after {\n  content: \"ab\";\n}\n", "")
	expectPrinted(t, "a:after { content: 'a\\\r\nb' }", "a:after {\n  content: \"ab\";\n}\n", "")
	expectPrinted(t, "a:after { content: 'a\\62 c' }", "a:after {\n  content: \"abc\";\n}\n", "")

	expectPrinted(t, "a:after { content: '\r' }", "a:after {\n  content: '\n  ' }\n  ;\n}\n",
		`<stdin>: WARNING: Unterminated string token
<stdin>: WARNING: Expected "}" to go with "{"
<stdin>: NOTE: The unbalanced "{" is here:
<stdin>: WARNING: Unterminated string token
`)
	expectPrinted(t, "a:after { content: '\n' }", "a:after {\n  content: '\n  ' }\n  ;\n}\n",
		`<stdin>: WARNING: Unterminated string token
<stdin>: WARNING: Expected "}" to go with "{"
<stdin>: NOTE: The unbalanced "{" is here:
<stdin>: WARNING: Unterminated string token
`)
	expectPrinted(t, "a:after { content: '\f' }", "a:after {\n  content: '\n  ' }\n  ;\n}\n",
		`<stdin>: WARNING: Unterminated string token
<stdin>: WARNING: Expected "}" to go with "{"
<stdin>: NOTE: The unbalanced "{" is here:
<stdin>: WARNING: Unterminated string token
`)
	expectPrinted(t, "a:after { content: '\r\n' }", "a:after {\n  content: '\n  ' }\n  ;\n}\n",
		`<stdin>: WARNING: Unterminated string token
<stdin>: WARNING: Expected "}" to go with "{"
<stdin>: NOTE: The unbalanced "{" is here:
<stdin>: WARNING: Unterminated string token
`)

	expectPrinted(t, "a:after { content: '\\1010101' }", "a:after {\n  content: \"\U001010101\";\n}\n", "")
	expectPrinted(t, "a:after { content: '\\invalid' }", "a:after {\n  content: \"invalid\";\n}\n", "")
}

func TestNumber(t *testing.T) {
	for _, ext := range []string{"", "%", "px+"} {
		expectPrinted(t, "a { width: .0"+ext+"; }", "a {\n  width: .0"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: .00"+ext+"; }", "a {\n  width: .00"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: .10"+ext+"; }", "a {\n  width: .10"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: 0."+ext+"; }", "a {\n  width: 0."+ext+";\n}\n", "")
		expectPrinted(t, "a { width: 0.0"+ext+"; }", "a {\n  width: 0.0"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: 0.1"+ext+"; }", "a {\n  width: 0.1"+ext+";\n}\n", "")

		expectPrinted(t, "a { width: +.0"+ext+"; }", "a {\n  width: +.0"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: +.00"+ext+"; }", "a {\n  width: +.00"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: +.10"+ext+"; }", "a {\n  width: +.10"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: +0."+ext+"; }", "a {\n  width: +0."+ext+";\n}\n", "")
		expectPrinted(t, "a { width: +0.0"+ext+"; }", "a {\n  width: +0.0"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: +0.1"+ext+"; }", "a {\n  width: +0.1"+ext+";\n}\n", "")

		expectPrinted(t, "a { width: -.0"+ext+"; }", "a {\n  width: -.0"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: -.00"+ext+"; }", "a {\n  width: -.00"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: -.10"+ext+"; }", "a {\n  width: -.10"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: -0."+ext+"; }", "a {\n  width: -0."+ext+";\n}\n", "")
		expectPrinted(t, "a { width: -0.0"+ext+"; }", "a {\n  width: -0.0"+ext+";\n}\n", "")
		expectPrinted(t, "a { width: -0.1"+ext+"; }", "a {\n  width: -0.1"+ext+";\n}\n", "")

		expectPrintedMangle(t, "a { width: .0"+ext+"; }", "a {\n  width: 0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: .00"+ext+"; }", "a {\n  width: 0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: .10"+ext+"; }", "a {\n  width: .1"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: 0."+ext+"; }", "a {\n  width: 0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: 0.0"+ext+"; }", "a {\n  width: 0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: 0.1"+ext+"; }", "a {\n  width: .1"+ext+";\n}\n", "")

		expectPrintedMangle(t, "a { width: +.0"+ext+"; }", "a {\n  width: +0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: +.00"+ext+"; }", "a {\n  width: +0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: +.10"+ext+"; }", "a {\n  width: +.1"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: +0."+ext+"; }", "a {\n  width: +0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: +0.0"+ext+"; }", "a {\n  width: +0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: +0.1"+ext+"; }", "a {\n  width: +.1"+ext+";\n}\n", "")

		expectPrintedMangle(t, "a { width: -.0"+ext+"; }", "a {\n  width: -0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: -.00"+ext+"; }", "a {\n  width: -0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: -.10"+ext+"; }", "a {\n  width: -.1"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: -0."+ext+"; }", "a {\n  width: -0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: -0.0"+ext+"; }", "a {\n  width: -0"+ext+";\n}\n", "")
		expectPrintedMangle(t, "a { width: -0.1"+ext+"; }", "a {\n  width: -.1"+ext+";\n}\n", "")
	}
}

func TestURL(t *testing.T) {
	expectPrinted(t, "a { background: url(foo.png) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url('foo.png') }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\") }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\" ) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\"\t) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\"\r) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\"\n) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\"\f) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\"\r\n) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url( \"foo.png\") }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\t\"foo.png\") }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\r\"foo.png\") }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\n\"foo.png\") }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\f\"foo.png\") }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\r\n\"foo.png\") }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url( \"foo.png\" ) }", "a {\n  background: url(foo.png);\n}\n", "")
	expectPrinted(t, "a { background: url(\"foo.png\" extra-stuff) }", "a {\n  background: url(\"foo.png\" extra-stuff);\n}\n", "")
	expectPrinted(t, "a { background: url( \"foo.png\" extra-stuff ) }", "a {\n  background: url(\"foo.png\" extra-stuff);\n}\n", "")
}

func TestHexColor(t *testing.T) {
	// "#RGBA"

	expectPrinted(t, "a { color: #1234 }", "a {\n  color: #1234;\n}\n", "")
	expectPrinted(t, "a { color: #123f }", "a {\n  color: #123f;\n}\n", "")
	expectPrinted(t, "a { color: #abcd }", "a {\n  color: #abcd;\n}\n", "")
	expectPrinted(t, "a { color: #abcf }", "a {\n  color: #abcf;\n}\n", "")
	expectPrinted(t, "a { color: #ABCD }", "a {\n  color: #ABCD;\n}\n", "")
	expectPrinted(t, "a { color: #ABCF }", "a {\n  color: #ABCF;\n}\n", "")

	expectPrintedMangle(t, "a { color: #1234 }", "a {\n  color: #1234;\n}\n", "")
	expectPrintedMangle(t, "a { color: #123f }", "a {\n  color: #123;\n}\n", "")
	expectPrintedMangle(t, "a { color: #abcd }", "a {\n  color: #abcd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #abcf }", "a {\n  color: #abc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ABCD }", "a {\n  color: #abcd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ABCF }", "a {\n  color: #abc;\n}\n", "")

	// "#RRGGBB"

	expectPrinted(t, "a { color: #112233 }", "a {\n  color: #112233;\n}\n", "")
	expectPrinted(t, "a { color: #122233 }", "a {\n  color: #122233;\n}\n", "")
	expectPrinted(t, "a { color: #112333 }", "a {\n  color: #112333;\n}\n", "")
	expectPrinted(t, "a { color: #112234 }", "a {\n  color: #112234;\n}\n", "")

	expectPrintedMangle(t, "a { color: #112233 }", "a {\n  color: #123;\n}\n", "")
	expectPrintedMangle(t, "a { color: #122233 }", "a {\n  color: #122233;\n}\n", "")
	expectPrintedMangle(t, "a { color: #112333 }", "a {\n  color: #112333;\n}\n", "")
	expectPrintedMangle(t, "a { color: #112234 }", "a {\n  color: #112234;\n}\n", "")

	expectPrinted(t, "a { color: #aabbcc }", "a {\n  color: #aabbcc;\n}\n", "")
	expectPrinted(t, "a { color: #abbbcc }", "a {\n  color: #abbbcc;\n}\n", "")
	expectPrinted(t, "a { color: #aabccc }", "a {\n  color: #aabccc;\n}\n", "")
	expectPrinted(t, "a { color: #aabbcd }", "a {\n  color: #aabbcd;\n}\n", "")

	expectPrintedMangle(t, "a { color: #aabbcc }", "a {\n  color: #abc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #abbbcc }", "a {\n  color: #abbbcc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabccc }", "a {\n  color: #aabccc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabbcd }", "a {\n  color: #aabbcd;\n}\n", "")

	expectPrinted(t, "a { color: #AABBCC }", "a {\n  color: #AABBCC;\n}\n", "")
	expectPrinted(t, "a { color: #ABBBCC }", "a {\n  color: #ABBBCC;\n}\n", "")
	expectPrinted(t, "a { color: #AABCCC }", "a {\n  color: #AABCCC;\n}\n", "")
	expectPrinted(t, "a { color: #AABBCD }", "a {\n  color: #AABBCD;\n}\n", "")

	expectPrintedMangle(t, "a { color: #AABBCC }", "a {\n  color: #abc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ABBBCC }", "a {\n  color: #abbbcc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABCCC }", "a {\n  color: #aabccc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABBCD }", "a {\n  color: #aabbcd;\n}\n", "")

	// "#RRGGBBAA"

	expectPrinted(t, "a { color: #11223344 }", "a {\n  color: #11223344;\n}\n", "")
	expectPrinted(t, "a { color: #12223344 }", "a {\n  color: #12223344;\n}\n", "")
	expectPrinted(t, "a { color: #11233344 }", "a {\n  color: #11233344;\n}\n", "")
	expectPrinted(t, "a { color: #11223444 }", "a {\n  color: #11223444;\n}\n", "")
	expectPrinted(t, "a { color: #11223345 }", "a {\n  color: #11223345;\n}\n", "")

	expectPrintedMangle(t, "a { color: #11223344 }", "a {\n  color: #1234;\n}\n", "")
	expectPrintedMangle(t, "a { color: #12223344 }", "a {\n  color: #12223344;\n}\n", "")
	expectPrintedMangle(t, "a { color: #11233344 }", "a {\n  color: #11233344;\n}\n", "")
	expectPrintedMangle(t, "a { color: #11223444 }", "a {\n  color: #11223444;\n}\n", "")
	expectPrintedMangle(t, "a { color: #11223345 }", "a {\n  color: #11223345;\n}\n", "")

	expectPrinted(t, "a { color: #aabbccdd }", "a {\n  color: #aabbccdd;\n}\n", "")
	expectPrinted(t, "a { color: #abbbccdd }", "a {\n  color: #abbbccdd;\n}\n", "")
	expectPrinted(t, "a { color: #aabcccdd }", "a {\n  color: #aabcccdd;\n}\n", "")
	expectPrinted(t, "a { color: #aabbcddd }", "a {\n  color: #aabbcddd;\n}\n", "")
	expectPrinted(t, "a { color: #aabbccde }", "a {\n  color: #aabbccde;\n}\n", "")

	expectPrintedMangle(t, "a { color: #aabbccdd }", "a {\n  color: #abcd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #abbbccdd }", "a {\n  color: #abbbccdd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabcccdd }", "a {\n  color: #aabcccdd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabbcddd }", "a {\n  color: #aabbcddd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabbccde }", "a {\n  color: #aabbccde;\n}\n", "")

	expectPrinted(t, "a { color: #AABBCCDD }", "a {\n  color: #AABBCCDD;\n}\n", "")
	expectPrinted(t, "a { color: #ABBBCCDD }", "a {\n  color: #ABBBCCDD;\n}\n", "")
	expectPrinted(t, "a { color: #AABCCCDD }", "a {\n  color: #AABCCCDD;\n}\n", "")
	expectPrinted(t, "a { color: #AABBCDDD }", "a {\n  color: #AABBCDDD;\n}\n", "")
	expectPrinted(t, "a { color: #AABBCCDE }", "a {\n  color: #AABBCCDE;\n}\n", "")

	expectPrintedMangle(t, "a { color: #AABBCCDD }", "a {\n  color: #abcd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ABBBCCDD }", "a {\n  color: #abbbccdd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABCCCDD }", "a {\n  color: #aabcccdd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABBCDDD }", "a {\n  color: #aabbcddd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABBCCDE }", "a {\n  color: #aabbccde;\n}\n", "")

	// "#RRGGBBFF"

	expectPrinted(t, "a { color: #112233ff }", "a {\n  color: #112233ff;\n}\n", "")
	expectPrinted(t, "a { color: #122233ff }", "a {\n  color: #122233ff;\n}\n", "")
	expectPrinted(t, "a { color: #112333ff }", "a {\n  color: #112333ff;\n}\n", "")
	expectPrinted(t, "a { color: #112234ff }", "a {\n  color: #112234ff;\n}\n", "")
	expectPrinted(t, "a { color: #112233ef }", "a {\n  color: #112233ef;\n}\n", "")

	expectPrintedMangle(t, "a { color: #112233ff }", "a {\n  color: #123;\n}\n", "")
	expectPrintedMangle(t, "a { color: #122233ff }", "a {\n  color: #122233;\n}\n", "")
	expectPrintedMangle(t, "a { color: #112333ff }", "a {\n  color: #112333;\n}\n", "")
	expectPrintedMangle(t, "a { color: #112234ff }", "a {\n  color: #112234;\n}\n", "")
	expectPrintedMangle(t, "a { color: #112233ef }", "a {\n  color: #112233ef;\n}\n", "")

	expectPrinted(t, "a { color: #aabbccff }", "a {\n  color: #aabbccff;\n}\n", "")
	expectPrinted(t, "a { color: #abbbccff }", "a {\n  color: #abbbccff;\n}\n", "")
	expectPrinted(t, "a { color: #aabcccff }", "a {\n  color: #aabcccff;\n}\n", "")
	expectPrinted(t, "a { color: #aabbcdff }", "a {\n  color: #aabbcdff;\n}\n", "")
	expectPrinted(t, "a { color: #aabbccef }", "a {\n  color: #aabbccef;\n}\n", "")

	expectPrintedMangle(t, "a { color: #aabbccff }", "a {\n  color: #abc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #abbbccff }", "a {\n  color: #abbbcc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabcccff }", "a {\n  color: #aabccc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabbcdff }", "a {\n  color: #aabbcd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #aabbccef }", "a {\n  color: #aabbccef;\n}\n", "")

	expectPrinted(t, "a { color: #AABBCCFF }", "a {\n  color: #AABBCCFF;\n}\n", "")
	expectPrinted(t, "a { color: #ABBBCCFF }", "a {\n  color: #ABBBCCFF;\n}\n", "")
	expectPrinted(t, "a { color: #AABCCCFF }", "a {\n  color: #AABCCCFF;\n}\n", "")
	expectPrinted(t, "a { color: #AABBCDFF }", "a {\n  color: #AABBCDFF;\n}\n", "")
	expectPrinted(t, "a { color: #AABBCCEF }", "a {\n  color: #AABBCCEF;\n}\n", "")

	expectPrintedMangle(t, "a { color: #AABBCCFF }", "a {\n  color: #abc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ABBBCCFF }", "a {\n  color: #abbbcc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABCCCFF }", "a {\n  color: #aabccc;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABBCDFF }", "a {\n  color: #aabbcd;\n}\n", "")
	expectPrintedMangle(t, "a { color: #AABBCCEF }", "a {\n  color: #aabbccef;\n}\n", "")
}

func TestColorFunctions(t *testing.T) {
	expectPrinted(t, "a { color: color(display-p3 0.5 0.0 0.0%) }", "a {\n  color: color(display-p3 0.5 0.0 0.0%);\n}\n", "")
	expectPrinted(t, "a { color: color(display-p3 0.5 0.0 0.0% / 0.5) }", "a {\n  color: color(display-p3 0.5 0.0 0.0% / 0.5);\n}\n", "")

	// Check minification of tokens
	expectPrintedMangle(t, "a { color: color(display-p3 0.5 0.0 0.0%) }", "a {\n  color: color(display-p3 .5 0 0%);\n}\n", "")
	expectPrintedMangle(t, "a { color: color(display-p3 0.5 0.0 0.0% / 0.5) }", "a {\n  color: color(display-p3 .5 0 0% / .5);\n}\n", "")

	// Check out-of-range colors
	expectPrintedLower(t, "a { before: 0; color: color(display-p3 1 0 0); after: 1 }",
		"a {\n  before: 0;\n  color: #ff0f0e;\n  color: color(display-p3 1 0 0);\n  after: 1;\n}\n", "")
	expectPrintedLowerMangle(t, "a { before: 0; color: color(display-p3 1 0 0); after: 1 }",
		"a {\n  before: 0;\n  color: #ff0f0e;\n  color: color(display-p3 1 0 0);\n  after: 1;\n}\n", "")
	expectPrintedLower(t, "a { before: 0; color: color(display-p3 1 0 0 / 0.5); after: 1 }",
		"a {\n  before: 0;\n  color: rgba(255, 15, 14, .5);\n  color: color(display-p3 1 0 0 / 0.5);\n  after: 1;\n}\n", "")
	expectPrintedLowerMangle(t, "a { before: 0; color: color(display-p3 1 0 0 / 0.5); after: 1 }",
		"a {\n  before: 0;\n  color: rgba(255, 15, 14, .5);\n  color: color(display-p3 1 0 0 / .5);\n  after: 1;\n}\n", "")
	expectPrintedLower(t, "a { before: 0; background: color(display-p3 1 0 0); after: 1 }",
		"a {\n  before: 0;\n  background: #ff0f0e;\n  background: color(display-p3 1 0 0);\n  after: 1;\n}\n", "")
	expectPrintedLowerMangle(t, "a { before: 0; background: color(display-p3 1 0 0); after: 1 }",
		"a {\n  before: 0;\n  background: #ff0f0e;\n  background: color(display-p3 1 0 0);\n  after: 1;\n}\n", "")
	expectPrintedLower(t, "a { before: 0; background: color(display-p3 1 0 0 / 0.5); after: 1 }",
		"a {\n  before: 0;\n  background: rgba(255, 15, 14, .5);\n  background: color(display-p3 1 0 0 / 0.5);\n  after: 1;\n}\n", "")
	expectPrintedLowerMangle(t, "a { before: 0; background: color(display-p3 1 0 0 / 0.5); after: 1 }",
		"a {\n  before: 0;\n  background: rgba(255, 15, 14, .5);\n  background: color(display-p3 1 0 0 / .5);\n  after: 1;\n}\n", "")
	expectPrintedLower(t, "a { before: 0; box-shadow: 1px color(display-p3 1 0 0); after: 1 }",
		"a {\n  before: 0;\n  box-shadow: 1px #ff0f0e;\n  box-shadow: 1px color(display-p3 1 0 0);\n  after: 1;\n}\n", "")
	expectPrintedLowerMangle(t, "a { before: 0; box-shadow: 1px color(display-p3 1 0 0); after: 1 }",
		"a {\n  before: 0;\n  box-shadow: 1px #ff0f0e;\n  box-shadow: 1px color(display-p3 1 0 0);\n  after: 1;\n}\n", "")
	expectPrintedLower(t, "a { before: 0; box-shadow: 1px color(display-p3 1 0 0 / 0.5); after: 1 }",
		"a {\n  before: 0;\n  box-shadow: 1px rgba(255, 15, 14, .5);\n  box-shadow: 1px color(display-p3 1 0 0 / 0.5);\n  after: 1;\n}\n", "")
	expectPrintedLowerMangle(t, "a { before: 0; box-shadow: 1px color(display-p3 1 0 0 / 0.5); after: 1 }",
		"a {\n  before: 0;\n  box-shadow: 1px rgba(255, 15, 14, .5);\n  box-shadow: 1px color(display-p3 1 0 0 / .5);\n  after: 1;\n}\n", "")

	// Don't insert a fallback after a previous instance of the same property
	expectPrintedLower(t, "a { color: red; color: color(display-p3 1 0 0) }",
		"a {\n  color: red;\n  color: color(display-p3 1 0 0);\n}\n", "")
	expectPrintedLower(t, "a { color: color(display-p3 1 0 0); color: color(display-p3 0 1 0) }",
		"a {\n  color: #ff0f0e;\n  color: color(display-p3 1 0 0);\n  color: color(display-p3 0 1 0);\n}\n", "")

	// Check case sensitivity
	expectPrintedLower(t, "a { color: color(srgb 0.87 0.98 0.807) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "A { Color: Color(Srgb 0.87 0.98 0.807) }", "A {\n  Color: #deface;\n}\n", "")
	expectPrintedLower(t, "A { COLOR: COLOR(SRGB 0.87 0.98 0.807) }", "A {\n  COLOR: #deface;\n}\n", "")

	// Check in-range colors in various color spaces
	expectPrintedLower(t, "a { color: color(a98-rgb 0.9 0.98 0.81) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(a98-rgb 90% 98% 81%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(display-p3 0.89 0.977 0.823) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(display-p3 89% 97.7% 82.3%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(prophoto-rgb 0.877 0.959 0.793) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(prophoto-rgb 87.7% 95.9% 79.3%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(rec2020 0.895 0.968 0.805) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(rec2020 89.5% 96.8% 80.5%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(srgb 0.87 0.98 0.807) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(srgb 87% 98% 80.7%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(srgb-linear 0.73 0.96 0.62) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(srgb-linear 73% 96% 62%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(xyz 0.754 0.883 0.715) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(xyz 75.4% 88.3% 71.5%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(xyz-d50 0.773 0.883 0.545) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(xyz-d50 77.3% 88.3% 54.5%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(xyz-d65 0.754 0.883 0.715) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: color(xyz-d65 75.4% 88.3% 71.5%) }", "a {\n  color: #deface;\n}\n", "")

	// Check color functions with unusual percent reference ranges
	expectPrintedLower(t, "a { color: lab(95.38 -15 18) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: lab(95.38% -15 18) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: lab(95.38 -12% 18) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: lab(95.38% -15 14.4%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: lch(95.38 23.57 130.22) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: lch(95.38% 23.57 130.22) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: lch(95.38 19% 130.22) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: lch(95.38 23.57 0.362turn) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklab(0.953 -0.045 0.046) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklab(95.3% -0.045 0.046) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklab(0.953 -11.2% 0.046) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklab(0.953 -0.045 11.5%) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklch(0.953 0.064 134) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklch(95.3% 0.064 134) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklch(0.953 16% 134) }", "a {\n  color: #deface;\n}\n", "")
	expectPrintedLower(t, "a { color: oklch(0.953 0.064 0.372turn) }", "a {\n  color: #deface;\n}\n", "")

	// Test alpha
	expectPrintedLower(t, "a { color: color(srgb 0.87 0.98 0.807 / 0.5) }", "a {\n  color: rgba(222, 250, 206, .5);\n}\n", "")
	expectPrintedLower(t, "a { color: lab(95.38 -15 18 / 0.5) }", "a {\n  color: rgba(222, 250, 206, .5);\n}\n", "")
	expectPrintedLower(t, "a { color: lch(95.38 23.57 130.22 / 0.5) }", "a {\n  color: rgba(222, 250, 206, .5);\n}\n", "")
	expectPrintedLower(t, "a { color: oklab(0.953 -0.045 0.046 / 0.5) }", "a {\n  color: rgba(222, 250, 206, .5);\n}\n", "")
	expectPrintedLower(t, "a { color: oklch(0.953 0.064 134 / 0.5) }", "a {\n  color: rgba(222, 250, 206, .5);\n}\n", "")
}

func TestColorNames(t *testing.T) {
	expectPrinted(t, "a { color: #f00 }", "a {\n  color: #f00;\n}\n", "")
	expectPrinted(t, "a { color: #f00f }", "a {\n  color: #f00f;\n}\n", "")
	expectPrinted(t, "a { color: #ff0000 }", "a {\n  color: #ff0000;\n}\n", "")
	expectPrinted(t, "a { color: #ff0000ff }", "a {\n  color: #ff0000ff;\n}\n", "")

	expectPrintedMangle(t, "a { color: #f00 }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: #f00e }", "a {\n  color: #f00e;\n}\n", "")
	expectPrintedMangle(t, "a { color: #f00f }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ff0000 }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ff0000ef }", "a {\n  color: #ff0000ef;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ff0000ff }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ffc0cb }", "a {\n  color: pink;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ffc0cbef }", "a {\n  color: #ffc0cbef;\n}\n", "")
	expectPrintedMangle(t, "a { color: #ffc0cbff }", "a {\n  color: pink;\n}\n", "")

	expectPrinted(t, "a { color: white }", "a {\n  color: white;\n}\n", "")
	expectPrinted(t, "a { color: tUrQuOiSe }", "a {\n  color: tUrQuOiSe;\n}\n", "")

	expectPrintedMangle(t, "a { color: white }", "a {\n  color: #fff;\n}\n", "")
	expectPrintedMangle(t, "a { color: tUrQuOiSe }", "a {\n  color: #40e0d0;\n}\n", "")
}

func TestColorRGBA(t *testing.T) {
	expectPrintedMangle(t, "a { color: rgba(1 2 3 / 0.5) }", "a {\n  color: #01020380;\n}\n", "")
	expectPrintedMangle(t, "a { color: rgba(1 2 3 / 50%) }", "a {\n  color: #0102037f;\n}\n", "")
	expectPrintedMangle(t, "a { color: rgba(1, 2, 3, 0.5) }", "a {\n  color: #01020380;\n}\n", "")
	expectPrintedMangle(t, "a { color: rgba(1, 2, 3, 50%) }", "a {\n  color: #0102037f;\n}\n", "")
	expectPrintedMangle(t, "a { color: rgba(1% 2% 3% / 0.5) }", "a {\n  color: #03050880;\n}\n", "")
	expectPrintedMangle(t, "a { color: rgba(1% 2% 3% / 50%) }", "a {\n  color: #0305087f;\n}\n", "")
	expectPrintedMangle(t, "a { color: rgba(1%, 2%, 3%, 0.5) }", "a {\n  color: #03050880;\n}\n", "")
	expectPrintedMangle(t, "a { color: rgba(1%, 2%, 3%, 50%) }", "a {\n  color: #0305087f;\n}\n", "")

	expectPrintedLowerMangle(t, "a { color: rgb(1, 2, 3, 0.4) }", "a {\n  color: rgba(1, 2, 3, .4);\n}\n", "")
	expectPrintedLowerMangle(t, "a { color: rgba(1, 2, 3, 40%) }", "a {\n  color: rgba(1, 2, 3, .4);\n}\n", "")

	expectPrintedLowerMangle(t, "a { color: rgb(var(--x) var(--y) var(--z)) }", "a {\n  color: rgb(var(--x) var(--y) var(--z));\n}\n", "")
}

func TestColorHSLA(t *testing.T) {
	expectPrintedMangle(t, ".red { color: hsl(0, 100%, 50%) }", ".red {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, ".orange { color: hsl(30deg, 100%, 50%) }", ".orange {\n  color: #ff8000;\n}\n", "")
	expectPrintedMangle(t, ".yellow { color: hsl(60 100% 50%) }", ".yellow {\n  color: #ff0;\n}\n", "")
	expectPrintedMangle(t, ".green { color: hsl(120, 100%, 50%) }", ".green {\n  color: #0f0;\n}\n", "")
	expectPrintedMangle(t, ".cyan { color: hsl(200grad, 100%, 50%) }", ".cyan {\n  color: #0ff;\n}\n", "")
	expectPrintedMangle(t, ".blue { color: hsl(240, 100%, 50%) }", ".blue {\n  color: #00f;\n}\n", "")
	expectPrintedMangle(t, ".purple { color: hsl(0.75turn 100% 50%) }", ".purple {\n  color: #7f00ff;\n}\n", "")
	expectPrintedMangle(t, ".magenta { color: hsl(300, 100%, 50%) }", ".magenta {\n  color: #f0f;\n}\n", "")

	expectPrintedMangle(t, "a { color: hsl(30 25% 50% / 50%) }", "a {\n  color: #9f80607f;\n}\n", "")
	expectPrintedMangle(t, "a { color: hsla(30 25% 50% / 50%) }", "a {\n  color: #9f80607f;\n}\n", "")

	expectPrintedLowerMangle(t, "a { color: hsl(1, 2%, 3%, 0.4) }", "a {\n  color: rgba(8, 8, 7, .4);\n}\n", "")
	expectPrintedLowerMangle(t, "a { color: hsla(1, 2%, 3%, 40%) }", "a {\n  color: rgba(8, 8, 7, .4);\n}\n", "")

	expectPrintedLowerMangle(t, "a { color: hsl(var(--x) var(--y) var(--z)) }", "a {\n  color: hsl(var(--x) var(--y) var(--z));\n}\n", "")
}

func TestLowerColor(t *testing.T) {
	expectPrintedLower(t, "a { color: rebeccapurple }", "a {\n  color: #663399;\n}\n", "")
	expectPrintedLower(t, "a { color: ReBeCcApUrPlE }", "a {\n  color: #663399;\n}\n", "")

	expectPrintedLower(t, "a { color: #0123 }", "a {\n  color: rgba(0, 17, 34, .2);\n}\n", "")
	expectPrintedLower(t, "a { color: #1230 }", "a {\n  color: rgba(17, 34, 51, 0);\n}\n", "")
	expectPrintedLower(t, "a { color: #1234 }", "a {\n  color: rgba(17, 34, 51, .267);\n}\n", "")
	expectPrintedLower(t, "a { color: #123f }", "a {\n  color: #112233;\n}\n", "")
	expectPrintedLower(t, "a { color: #12345678 }", "a {\n  color: rgba(18, 52, 86, .47);\n}\n", "")
	expectPrintedLower(t, "a { color: #ff00007f }", "a {\n  color: rgba(255, 0, 0, .498);\n}\n", "")

	expectPrintedLower(t, "a { color: rgb(1 2 3) }", "a {\n  color: rgb(1, 2, 3);\n}\n", "")
	expectPrintedLower(t, "a { color: hsl(1 2% 3%) }", "a {\n  color: hsl(1, 2%, 3%);\n}\n", "")
	expectPrintedLower(t, "a { color: rgba(1% 2% 3%) }", "a {\n  color: rgb(1%, 2%, 3%);\n}\n", "")
	expectPrintedLower(t, "a { color: hsla(1deg 2% 3%) }", "a {\n  color: hsl(1, 2%, 3%);\n}\n", "")

	expectPrintedLower(t, "a { color: hsla(200grad 2% 3%) }", "a {\n  color: hsl(180, 2%, 3%);\n}\n", "")
	expectPrintedLower(t, "a { color: hsla(6.28319rad 2% 3%) }", "a {\n  color: hsl(360, 2%, 3%);\n}\n", "")
	expectPrintedLower(t, "a { color: hsla(0.5turn 2% 3%) }", "a {\n  color: hsl(180, 2%, 3%);\n}\n", "")
	expectPrintedLower(t, "a { color: hsla(+200grad 2% 3%) }", "a {\n  color: hsl(180, 2%, 3%);\n}\n", "")
	expectPrintedLower(t, "a { color: hsla(-200grad 2% 3%) }", "a {\n  color: hsl(-180, 2%, 3%);\n}\n", "")

	expectPrintedLower(t, "a { color: rgb(1 2 3 / 4) }", "a {\n  color: rgba(1, 2, 3, 4);\n}\n", "")
	expectPrintedLower(t, "a { color: RGB(1 2 3 / 4) }", "a {\n  color: rgba(1, 2, 3, 4);\n}\n", "")
	expectPrintedLower(t, "a { color: rgba(1% 2% 3% / 4%) }", "a {\n  color: rgba(1%, 2%, 3%, 0.04);\n}\n", "")
	expectPrintedLower(t, "a { color: RGBA(1% 2% 3% / 4%) }", "a {\n  color: RGBA(1%, 2%, 3%, 0.04);\n}\n", "")
	expectPrintedLower(t, "a { color: hsl(1 2% 3% / 4) }", "a {\n  color: hsla(1, 2%, 3%, 4);\n}\n", "")
	expectPrintedLower(t, "a { color: HSL(1 2% 3% / 4) }", "a {\n  color: hsla(1, 2%, 3%, 4);\n}\n", "")
	expectPrintedLower(t, "a { color: hsla(1 2% 3% / 4%) }", "a {\n  color: hsla(1, 2%, 3%, 0.04);\n}\n", "")
	expectPrintedLower(t, "a { color: HSLA(1 2% 3% / 4%) }", "a {\n  color: HSLA(1, 2%, 3%, 0.04);\n}\n", "")

	expectPrintedLower(t, "a { color: rgb(1, 2, 3, 4) }", "a {\n  color: rgba(1, 2, 3, 4);\n}\n", "")
	expectPrintedLower(t, "a { color: rgba(1%, 2%, 3%, 4%) }", "a {\n  color: rgba(1%, 2%, 3%, 0.04);\n}\n", "")
	expectPrintedLower(t, "a { color: rgb(1%, 2%, 3%, 0.4%) }", "a {\n  color: rgba(1%, 2%, 3%, 0.004);\n}\n", "")

	expectPrintedLower(t, "a { color: hsl(1, 2%, 3%, 4) }", "a {\n  color: hsla(1, 2%, 3%, 4);\n}\n", "")
	expectPrintedLower(t, "a { color: hsla(1deg, 2%, 3%, 4%) }", "a {\n  color: hsla(1, 2%, 3%, 0.04);\n}\n", "")
	expectPrintedLower(t, "a { color: hsl(1deg, 2%, 3%, 0.4%) }", "a {\n  color: hsla(1, 2%, 3%, 0.004);\n}\n", "")

	expectPrintedLower(t, "a { color: hwb(90deg 20% 40%) }", "a {\n  color: #669933;\n}\n", "")
	expectPrintedLower(t, "a { color: HWB(90deg 20% 40%) }", "a {\n  color: #669933;\n}\n", "")
	expectPrintedLower(t, "a { color: hwb(90deg 20% 40% / 0.2) }", "a {\n  color: rgba(102, 153, 51, .2);\n}\n", "")
	expectPrintedLower(t, "a { color: hwb(1deg 40% 80%) }", "a {\n  color: #555555;\n}\n", "")
	expectPrintedLower(t, "a { color: hwb(1deg 9000% 50%) }", "a {\n  color: #aaaaaa;\n}\n", "")
	expectPrintedLower(t, "a { color: hwb(1deg 9000% 50% / 0.6) }", "a {\n  color: rgba(170, 170, 170, .6);\n}\n", "")
	expectPrintedLower(t, "a { color: hwb(90deg, 20%, 40%) }", "a {\n  color: hwb(90deg, 20%, 40%);\n}\n", "") // This is invalid
	expectPrintedLower(t, "a { color: hwb(none 20% 40%) }", "a {\n  color: hwb(none 20% 40%);\n}\n", "")       // Interpolation
	expectPrintedLower(t, "a { color: hwb(90deg none 40%) }", "a {\n  color: hwb(90deg none 40%);\n}\n", "")   // Interpolation
	expectPrintedLower(t, "a { color: hwb(90deg 20% none) }", "a {\n  color: hwb(90deg 20% none);\n}\n", "")   // Interpolation

	expectPrintedMangle(t, "a { color: hwb(90deg 20% 40%) }", "a {\n  color: #693;\n}\n", "")
	expectPrintedMangle(t, "a { color: hwb(0.75turn 20% 40% / 0.75) }", "a {\n  color: #663399bf;\n}\n", "")
	expectPrintedLowerMangle(t, "a { color: hwb(90deg 20% 40%) }", "a {\n  color: #693;\n}\n", "")
	expectPrintedLowerMangle(t, "a { color: hwb(0.75turn 20% 40% / 0.75) }", "a {\n  color: rgba(102, 51, 153, .75);\n}\n", "")
}

func TestBackground(t *testing.T) {
	expectPrinted(t, "a { background: #11223344 }", "a {\n  background: #11223344;\n}\n", "")
	expectPrintedMangle(t, "a { background: #11223344 }", "a {\n  background: #1234;\n}\n", "")
	expectPrintedLower(t, "a { background: #11223344 }", "a {\n  background: rgba(17, 34, 51, .267);\n}\n", "")

	expectPrinted(t, "a { background: border-box #11223344 }", "a {\n  background: border-box #11223344;\n}\n", "")
	expectPrintedMangle(t, "a { background: border-box #11223344 }", "a {\n  background: border-box #1234;\n}\n", "")
	expectPrintedLower(t, "a { background: border-box #11223344 }", "a {\n  background: border-box rgba(17, 34, 51, .267);\n}\n", "")
}

func TestGradient(t *testing.T) {
	gradientKinds := []string{
		"linear-gradient",
		"radial-gradient",
		"conic-gradient",
		"repeating-linear-gradient",
		"repeating-radial-gradient",
		"repeating-conic-gradient",
	}

	for _, gradient := range gradientKinds {
		var code string

		// Different properties
		expectPrinted(t, "a { background: "+gradient+"(red, blue) }", "a {\n  background: "+gradient+"(red, blue);\n}\n", "")
		expectPrinted(t, "a { background-image: "+gradient+"(red, blue) }", "a {\n  background-image: "+gradient+"(red, blue);\n}\n", "")
		expectPrinted(t, "a { border-image: "+gradient+"(red, blue) }", "a {\n  border-image: "+gradient+"(red, blue);\n}\n", "")
		expectPrinted(t, "a { mask-image: "+gradient+"(red, blue) }", "a {\n  mask-image: "+gradient+"(red, blue);\n}\n", "")

		// Basic
		code = "a { background: " + gradient + "(yellow, #11223344) }"
		expectPrinted(t, code, "a {\n  background: "+gradient+"(yellow, #11223344);\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background: "+gradient+"(#ff0, #1234);\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+"(yellow,#11223344)}", "")
		expectPrintedLowerUnsupported(t, compat.HexRGBA, code,
			"a {\n  background: "+gradient+"(yellow, rgba(17, 34, 51, .267));\n}\n", "")

		// Basic with positions
		code = "a { background: " + gradient + "(yellow 10%, #11223344 90%) }"
		expectPrinted(t, code, "a {\n  background: "+gradient+"(yellow 10%, #11223344 90%);\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background: "+gradient+"(#ff0 10%, #1234 90%);\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+"(yellow 10%,#11223344 90%)}", "")
		expectPrintedLowerUnsupported(t, compat.HexRGBA, code,
			"a {\n  background: "+gradient+"(yellow 10%, rgba(17, 34, 51, .267) 90%);\n}\n", "")

		// Basic with hints
		code = "a { background: " + gradient + "(yellow, 25%, #11223344) }"
		expectPrinted(t, code, "a {\n  background:\n    "+gradient+"(\n      yellow,\n      25%,\n      #11223344);\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background:\n    "+gradient+"(\n      #ff0,\n      25%,\n      #1234);\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+"(yellow,25%,#11223344)}", "")
		expectPrintedLowerUnsupported(t, compat.HexRGBA, code,
			"a {\n  background:\n    "+gradient+"(\n      yellow,\n      25%,\n      rgba(17, 34, 51, .267));\n}\n", "")
		expectPrintedLowerUnsupported(t, compat.GradientMidpoints, code,
			"a {\n  background:\n    "+gradient+"(\n      #ffff00,\n      #f2f303de,\n      #eced04d0 6.25%,\n      "+
				"#e1e306bd 12.5%,\n      #cdd00ba2 25%,\n      #a2a8147b,\n      #6873205d,\n      #11223344);\n}\n", "")

		// Double positions
		code = "a { background: " + gradient + "(green, red 10%, red 20%, yellow 70% 80%, black) }"
		expectPrinted(t, code, "a {\n  background:\n    "+gradient+"(\n      green,\n      red 10%,\n      "+
			"red 20%,\n      yellow 70% 80%,\n      black);\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background:\n    "+gradient+"(\n      green,\n      "+
			"red 10% 20%,\n      #ff0 70% 80%,\n      #000);\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+"(green,red 10%,red 20%,yellow 70% 80%,black)}", "")
		expectPrintedLowerUnsupported(t, compat.GradientDoublePosition, code,
			"a {\n  background:\n    "+gradient+"(\n      green,\n      red 10%,\n      red 20%,\n      "+
				"yellow 70%,\n      yellow 80%,\n      black);\n}\n", "")

		// Double positions with hints
		code = "a { background: " + gradient + "(green, red 10%, red 20%, 30%, yellow 70% 80%, 85%, black) }"
		expectPrinted(t, code, "a {\n  background:\n    "+gradient+"(\n      green,\n      red 10%,\n      red 20%,\n      "+
			"30%,\n      yellow 70% 80%,\n      85%,\n      black);\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background:\n    "+gradient+"(\n      green,\n      red 10% 20%,\n      "+
			"30%,\n      #ff0 70% 80%,\n      85%,\n      #000);\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+"(green,red 10%,red 20%,30%,yellow 70% 80%,85%,black)}", "")
		expectPrintedLowerUnsupported(t, compat.GradientDoublePosition, code,
			"a {\n  background:\n    "+gradient+"(\n      green,\n      red 10%,\n      red 20%,\n      30%,\n      "+
				"yellow 70%,\n      yellow 80%,\n      85%,\n      black);\n}\n", "")

		// Non-double positions with hints
		code = "a { background: " + gradient + "(green, red 10%, 1%, red 20%, black) }"
		expectPrinted(t, code, "a {\n  background:\n    "+gradient+"(\n      green,\n      red 10%,\n      1%,\n      "+
			"red 20%,\n      black);\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background:\n    "+gradient+"(\n      green,\n      red 10%,\n      "+
			"1%,\n      red 20%,\n      #000);\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+"(green,red 10%,1%,red 20%,black)}", "")
		expectPrintedLowerUnsupported(t, compat.GradientDoublePosition, code,
			"a {\n  background:\n    "+gradient+"(\n      green,\n      red 10%,\n      1%,\n      red 20%,\n      black);\n}\n", "")

		// Out-of-gamut colors
		code = "a { background: " + gradient + "(yellow, color(display-p3 1 0 0)) }"
		expectPrinted(t, code, "a {\n  background: "+gradient+"(yellow, color(display-p3 1 0 0));\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background: "+gradient+"(#ff0, color(display-p3 1 0 0));\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+"(yellow,color(display-p3 1 0 0))}", "")
		expectPrintedLowerUnsupported(t, compat.ColorFunctions, code,
			"a {\n  background:\n    "+gradient+"(\n      #ffff00,\n      #ffe971,\n      #ffd472 25%,\n      "+
				"#ffab5f,\n      #ff7b45 75%,\n      #ff5e38 87.5%,\n      #ff5534,\n      #ff4c30,\n      "+
				"#ff412c,\n      #ff0e0e);\n  "+
				"background:\n    "+gradient+"(\n      #ffff00,\n      color(xyz 0.734 0.805 0.111),\n      "+
				"color(xyz 0.699 0.693 0.087) 25%,\n      color(xyz 0.627 0.501 0.048),\n      "+
				"color(xyz 0.556 0.348 0.019) 75%,\n      color(xyz 0.521 0.284 0.009) 87.5%,\n      "+
				"color(xyz 0.512 0.27 0.006),\n      color(xyz 0.504 0.256 0.004),\n      "+
				"color(xyz 0.495 0.242 0.002),\n      color(xyz 0.487 0.229 0));\n}\n", "")

		// Whitespace
		code = "a { background: " + gradient + "(color-mix(in lab,red,green)calc(1px)calc(2px),color-mix(in lab,blue,red)calc(98%)calc(99%)) }"
		expectPrinted(t, code, "a {\n  background: "+gradient+
			"(color-mix(in lab, red, green)calc(1px)calc(2px), color-mix(in lab, blue, red)calc(98%)calc(99%));\n}\n", "")
		expectPrintedMangle(t, code, "a {\n  background: "+gradient+
			"(color-mix(in lab, red, green) 1px 2px, color-mix(in lab, blue, red) 98% 99%);\n}\n", "")
		expectPrintedMinify(t, code, "a{background:"+gradient+
			"(color-mix(in lab,red,green)calc(1px)calc(2px),color-mix(in lab,blue,red)calc(98%)calc(99%))}", "")
		expectPrintedLowerUnsupported(t, compat.GradientDoublePosition, code, "a {\n  background:\n    "+gradient+
			"(\n      color-mix(in lab, red, green) calc(1px),\n      color-mix(in lab, red, green) calc(2px),"+
			"\n      color-mix(in lab, blue, red) calc(98%),\n      color-mix(in lab, blue, red) calc(99%));\n}\n", "")
		expectPrintedLowerMangle(t, code, "a {\n  background:\n    "+gradient+
			"(\n      color-mix(in lab, red, green) 1px,\n      color-mix(in lab, red, green) 2px,"+
			"\n      color-mix(in lab, blue, red) 98%,\n      color-mix(in lab, blue, red) 99%);\n}\n", "")

		// Color space interpolation
		expectPrintedLowerUnsupported(t, compat.GradientInterpolation,
			"a { background: "+gradient+"(in srgb, red, green) }",
			"a {\n  background: "+gradient+"(#ff0000, #008000);\n}\n", "")
		expectPrintedLowerUnsupported(t, compat.GradientInterpolation,
			"a { background: "+gradient+"(in srgb-linear, red, green) }",
			"a {\n  background:\n    "+gradient+"(\n      #ff0000,\n      #fb1300,\n      #f81f00 6.25%,\n      "+
				"#f02e00 12.5%,\n      #e14200 25%,\n      #bc5c00,\n      #897000 75%,\n      #637800 87.5%,\n      "+
				"#477c00 93.75%,\n      #317e00,\n      #008000);\n}\n", "")
		expectPrintedLowerUnsupported(t, compat.GradientInterpolation,
			"a { background: "+gradient+"(in lab, red, green) }",
			"a {\n  background:\n    "+gradient+"(\n      #ff0000,\n      color(xyz 0.396 0.211 0.019),\n      "+
				"color(xyz 0.38 0.209 0.02) 6.25%,\n      color(xyz 0.35 0.205 0.02) 12.5%,\n      "+
				"color(xyz 0.294 0.198 0.02) 25%,\n      color(xyz 0.2 0.183 0.022),\n      "+
				"color(xyz 0.129 0.168 0.024) 75%,\n      color(xyz 0.101 0.161 0.025) 87.5%,\n      "+
				"color(xyz 0.089 0.158 0.025) 93.75%,\n      color(xyz 0.083 0.156 0.025),\n      #008000);\n}\n", "")

		// Hue interpolation
		expectPrintedLowerUnsupported(t, compat.GradientInterpolation,
			"a { background: "+gradient+"(in hsl shorter hue, red, green) }",
			"a {\n  background:\n    "+gradient+"(\n      #ff0000,\n      #df7000,\n      "+
				"#bfbf00,\n      #50a000,\n      #008000);\n}\n", "")
		expectPrintedLowerUnsupported(t, compat.GradientInterpolation,
			"a { background: "+gradient+"(in hsl longer hue, red, green) }",
			"a {\n  background:\n    "+gradient+"(\n      #ff0000,\n      #ef0078,\n      "+
				"#df00df,\n      #6800cf,\n      #0000c0,\n      #0058b0,\n      "+
				"#00a0a0,\n      #009048,\n      #008000);\n}\n", "")
	}
}

func TestDeclaration(t *testing.T) {
	expectPrinted(t, ".decl {}", ".decl {\n}\n", "")
	expectPrinted(t, ".decl { a: b }", ".decl {\n  a: b;\n}\n", "")
	expectPrinted(t, ".decl { a: b; }", ".decl {\n  a: b;\n}\n", "")
	expectPrinted(t, ".decl { a: b; c: d }", ".decl {\n  a: b;\n  c: d;\n}\n", "")
	expectPrinted(t, ".decl { a: b; c: d; }", ".decl {\n  a: b;\n  c: d;\n}\n", "")
	expectPrinted(t, ".decl { a { b: c; } }", ".decl {\n  a {\n    b: c;\n  }\n}\n", "")
	expectPrinted(t, ".decl { & a { b: c; } }", ".decl {\n  & a {\n    b: c;\n  }\n}\n", "")

	// See http://browserhacks.com/
	expectPrinted(t, ".selector { (;property: value;); }", ".selector {\n  (;property: value;);\n}\n", "<stdin>: WARNING: Expected identifier but found \"(\"\n")
	expectPrinted(t, ".selector { [;property: value;]; }", ".selector {\n  [;property: value;];\n}\n", "<stdin>: WARNING: Expected identifier but found \"[\"\n")
	expectPrinted(t, ".selector, {}", ".selector, {\n}\n", "<stdin>: WARNING: Unexpected \"{\"\n")
	expectPrinted(t, ".selector\\ {}", ".selector\\  {\n}\n", "")
	expectPrinted(t, ".selector { property: value\\9; }", ".selector {\n  property: value\\\t;\n}\n", "")
	expectPrinted(t, "@media \\0screen\\,screen\\9 {}", "@media \uFFFDscreen\\,screen\\\t {\n}\n", "")
}

func TestSelector(t *testing.T) {
	expectPrinted(t, "a{}", "a {\n}\n", "")
	expectPrinted(t, "a {}", "a {\n}\n", "")
	expectPrinted(t, "a b {}", "a b {\n}\n", "")

	expectPrinted(t, "a/**/b {}", "ab {\n}\n", "<stdin>: WARNING: Unexpected \"b\"\n")
	expectPrinted(t, "a/**/.b {}", "a.b {\n}\n", "")
	expectPrinted(t, "a/**/:b {}", "a:b {\n}\n", "")
	expectPrinted(t, "a/**/[b] {}", "a[b] {\n}\n", "")
	expectPrinted(t, "a>/**/b {}", "a > b {\n}\n", "")
	expectPrinted(t, "a+/**/b {}", "a + b {\n}\n", "")
	expectPrinted(t, "a~/**/b {}", "a ~ b {\n}\n", "")

	expectPrinted(t, "[b]{}", "[b] {\n}\n", "")
	expectPrinted(t, "[b] {}", "[b] {\n}\n", "")
	expectPrinted(t, "a[b] {}", "a[b] {\n}\n", "")
	expectPrinted(t, "a [b] {}", "a [b] {\n}\n", "")
	expectPrinted(t, "[] {}", "[] {\n}\n", "<stdin>: WARNING: Expected identifier but found \"]\"\n")
	expectPrinted(t, "[b {}", "[b] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")
	expectPrinted(t, "[b]] {}", "[b]] {\n}\n", "<stdin>: WARNING: Unexpected \"]\"\n")
	expectPrinted(t, "a[b {}", "a[b] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")
	expectPrinted(t, "a[b]] {}", "a[b]] {\n}\n", "<stdin>: WARNING: Unexpected \"]\"\n")
	expectPrinted(t, "[b]a {}", "[b]a {\n}\n", "<stdin>: WARNING: Unexpected \"a\"\n")

	expectPrinted(t, "[|b]{}", "[b] {\n}\n", "") // "[|b]" is equivalent to "[b]"
	expectPrinted(t, "[*|b]{}", "[*|b] {\n}\n", "")
	expectPrinted(t, "[a|b]{}", "[a|b] {\n}\n", "")
	expectPrinted(t, "[a|b|=\"c\"]{}", "[a|b|=c] {\n}\n", "")
	expectPrinted(t, "[a|b |= \"c\"]{}", "[a|b|=c] {\n}\n", "")
	expectPrinted(t, "[a||b] {}", "[a||b] {\n}\n", "<stdin>: WARNING: Expected identifier but found \"|\"\n")
	expectPrinted(t, "[* | b] {}", "[* | b] {\n}\n", "<stdin>: WARNING: Expected \"|\" but found whitespace\n")
	expectPrinted(t, "[a | b] {}", "[a | b] {\n}\n", "<stdin>: WARNING: Expected \"=\" but found whitespace\n")

	expectPrinted(t, "[b=\"c\"] {}", "[b=c] {\n}\n", "")
	expectPrinted(t, "[b=\"c d\"] {}", "[b=\"c d\"] {\n}\n", "")
	expectPrinted(t, "[b=\"0c\"] {}", "[b=\"0c\"] {\n}\n", "")
	expectPrinted(t, "[b~=\"c\"] {}", "[b~=c] {\n}\n", "")
	expectPrinted(t, "[b^=\"c\"] {}", "[b^=c] {\n}\n", "")
	expectPrinted(t, "[b$=\"c\"] {}", "[b$=c] {\n}\n", "")
	expectPrinted(t, "[b*=\"c\"] {}", "[b*=c] {\n}\n", "")
	expectPrinted(t, "[b|=\"c\"] {}", "[b|=c] {\n}\n", "")
	expectPrinted(t, "[b?=\"c\"] {}", "[b?=\"c\"] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")

	expectPrinted(t, "[b = \"c\"] {}", "[b=c] {\n}\n", "")
	expectPrinted(t, "[b ~= \"c\"] {}", "[b~=c] {\n}\n", "")
	expectPrinted(t, "[b ^= \"c\"] {}", "[b^=c] {\n}\n", "")
	expectPrinted(t, "[b $= \"c\"] {}", "[b$=c] {\n}\n", "")
	expectPrinted(t, "[b *= \"c\"] {}", "[b*=c] {\n}\n", "")
	expectPrinted(t, "[b |= \"c\"] {}", "[b|=c] {\n}\n", "")
	expectPrinted(t, "[b ?= \"c\"] {}", "[b ?= \"c\"] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")

	expectPrinted(t, "[b = \"c\" i] {}", "[b=c i] {\n}\n", "")
	expectPrinted(t, "[b = \"c\" I] {}", "[b=c I] {\n}\n", "")
	expectPrinted(t, "[b = \"c\" s] {}", "[b=c s] {\n}\n", "")
	expectPrinted(t, "[b = \"c\" S] {}", "[b=c S] {\n}\n", "")
	expectPrinted(t, "[b i] {}", "[b i] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")
	expectPrinted(t, "[b I] {}", "[b I] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")
	expectPrinted(t, "[b s] {}", "[b s] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")
	expectPrinted(t, "[b S] {}", "[b S] {\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")

	expectPrinted(t, "|b {}", "|b {\n}\n", "")
	expectPrinted(t, "|* {}", "|* {\n}\n", "")
	expectPrinted(t, "a|b {}", "a|b {\n}\n", "")
	expectPrinted(t, "a|* {}", "a|* {\n}\n", "")
	expectPrinted(t, "*|b {}", "*|b {\n}\n", "")
	expectPrinted(t, "*|* {}", "*|* {\n}\n", "")
	expectPrinted(t, "a||b {}", "a||b {\n}\n", "<stdin>: WARNING: Expected identifier but found \"|\"\n")

	expectPrinted(t, "a+b {}", "a + b {\n}\n", "")
	expectPrinted(t, "a>b {}", "a > b {\n}\n", "")
	expectPrinted(t, "a+b {}", "a + b {\n}\n", "")
	expectPrinted(t, "a~b {}", "a ~ b {\n}\n", "")

	expectPrinted(t, "a + b {}", "a + b {\n}\n", "")
	expectPrinted(t, "a > b {}", "a > b {\n}\n", "")
	expectPrinted(t, "a + b {}", "a + b {\n}\n", "")
	expectPrinted(t, "a ~ b {}", "a ~ b {\n}\n", "")

	expectPrinted(t, "::b {}", "::b {\n}\n", "")
	expectPrinted(t, "*::b {}", "*::b {\n}\n", "")
	expectPrinted(t, "a::b {}", "a::b {\n}\n", "")
	expectPrinted(t, "::b(c) {}", "::b(c) {\n}\n", "")
	expectPrinted(t, "*::b(c) {}", "*::b(c) {\n}\n", "")
	expectPrinted(t, "a::b(c) {}", "a::b(c) {\n}\n", "")
	expectPrinted(t, "a:b:c {}", "a:b:c {\n}\n", "")
	expectPrinted(t, "a:b(:c) {}", "a:b(:c) {\n}\n", "")
	expectPrinted(t, "a: b {}", "a: b {\n}\n", "<stdin>: WARNING: Expected identifier but found whitespace\n")
	expectPrinted(t, ":is(a)b {}", ":is(a)b {\n}\n", "<stdin>: WARNING: Unexpected \"b\"\n")
	expectPrinted(t, "a:b( c ) {}", "a:b(c) {\n}\n", "")
	expectPrinted(t, "a:b( c , d ) {}", "a:b(c, d) {\n}\n", "")
	expectPrinted(t, "a:is( c ) {}", "a:is(c) {\n}\n", "")
	expectPrinted(t, "a:is( c , d ) {}", "a:is(c, d) {\n}\n", "")

	// Check an empty <forgiving-selector-list> (see https://github.com/evanw/esbuild/issues/4232)
	expectPrinted(t, ":is() {}", ":is() {\n}\n", "")
	expectPrinted(t, ":where() {}", ":where() {\n}\n", "")
	expectPrinted(t, ":not(:is()) {}", ":not(:is()) {\n}\n", "")
	expectPrinted(t, ":not(:where()) {}", ":not(:where()) {\n}\n", "")
	expectPrinted(t, ":not() {}", ":not() {\n}\n", "<stdin>: WARNING: Unexpected \")\"\n")

	// These test cases previously caused a hang (see https://github.com/evanw/esbuild/issues/2276)
	expectPrinted(t, ":x(", ":x() {\n}\n", "<stdin>: WARNING: Unexpected end of file\n")
	expectPrinted(t, ":x( {}", ":x({}) {\n}\n", "<stdin>: WARNING: Expected \")\" to go with \"(\"\n<stdin>: NOTE: The unbalanced \"(\" is here:\n")
	expectPrinted(t, ":x(, :y() {}", ":x(, :y() {}) {\n}\n", "<stdin>: WARNING: Expected \")\" to go with \"(\"\n<stdin>: NOTE: The unbalanced \"(\" is here:\n")

	expectPrinted(t, "#id {}", "#id {\n}\n", "")
	expectPrinted(t, "#--0 {}", "#--0 {\n}\n", "")
	expectPrinted(t, "#\\-0 {}", "#\\-0 {\n}\n", "")
	expectPrinted(t, "#\\30 {}", "#\\30  {\n}\n", "")
	expectPrinted(t, "div#id {}", "div#id {\n}\n", "")
	expectPrinted(t, "div#--0 {}", "div#--0 {\n}\n", "")
	expectPrinted(t, "div#\\-0 {}", "div#\\-0 {\n}\n", "")
	expectPrinted(t, "div#\\30 {}", "div#\\30  {\n}\n", "")
	expectPrinted(t, "#0 {}", "#0 {\n}\n", "<stdin>: WARNING: Unexpected \"#0\"\n")
	expectPrinted(t, "#-0 {}", "#-0 {\n}\n", "<stdin>: WARNING: Unexpected \"#-0\"\n")
	expectPrinted(t, "div#0 {}", "div#0 {\n}\n", "<stdin>: WARNING: Unexpected \"#0\"\n")
	expectPrinted(t, "div#-0 {}", "div#-0 {\n}\n", "<stdin>: WARNING: Unexpected \"#-0\"\n")

	expectPrinted(t, "div::before::after::selection::first-line::first-letter {color:red}",
		"div::before::after::selection::first-line::first-letter {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "div::before::after::selection::first-line::first-letter {color:red}",
		"div:before:after::selection:first-line:first-letter {\n  color: red;\n}\n", "")

	// Make sure '-' and '\\' consume an ident-like token instead of a name
	expectPrinted(t, "_:-ms-lang(x) {}", "_:-ms-lang(x) {\n}\n", "")
	expectPrinted(t, "_:\\ms-lang(x) {}", "_:ms-lang(x) {\n}\n", "")

	expectPrinted(t, ":local(a, b) {}", ":local(a, b) {\n}\n", "<stdin>: WARNING: Unexpected \",\" inside \":local(...)\"\n"+
		"NOTE: Different CSS tools behave differently in this case, so esbuild doesn't allow it. Either remove "+
		"this comma or split this selector up into multiple comma-separated \":local(...)\" selectors instead.\n")
	expectPrinted(t, ":global(a, b) {}", ":global(a, b) {\n}\n", "<stdin>: WARNING: Unexpected \",\" inside \":global(...)\"\n"+
		"NOTE: Different CSS tools behave differently in this case, so esbuild doesn't allow it. Either remove "+
		"this comma or split this selector up into multiple comma-separated \":global(...)\" selectors instead.\n")
}

func TestNestedSelector(t *testing.T) {
	sassWarningWrap := "NOTE: CSS nesting syntax does not allow the \"&\" selector to come before " +
		"a type selector. You can wrap this selector in \":is(...)\" as a workaround. " +
		"This restriction exists to avoid problems with SASS nesting, where the same syntax " +
		"means something very different that has no equivalent in real CSS (appending a suffix to the parent selector).\n"
	sassWarningMove := "NOTE: CSS nesting syntax does not allow the \"&\" selector to come before " +
		"a type selector. You can move the \"&\" to the end of this selector as a workaround. " +
		"This restriction exists to avoid problems with SASS nesting, where the same syntax " +
		"means something very different that has no equivalent in real CSS (appending a suffix to the parent selector).\n"

	expectPrinted(t, "& {}", "& {\n}\n", "")
	expectPrinted(t, "& b {}", "& b {\n}\n", "")
	expectPrinted(t, "&:b {}", "&:b {\n}\n", "")
	expectPrinted(t, "&* {}", "&* {\n}\n", "<stdin>: WARNING: Cannot use type selector \"*\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "&|b {}", "&|b {\n}\n", "<stdin>: WARNING: Cannot use type selector \"|b\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "&*|b {}", "&*|b {\n}\n", "<stdin>: WARNING: Cannot use type selector \"*|b\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "&a|b {}", "&a|b {\n}\n", "<stdin>: WARNING: Cannot use type selector \"a|b\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "&[a] {}", "&[a] {\n}\n", "")

	expectPrinted(t, "a { & {} }", "a {\n  & {\n  }\n}\n", "")
	expectPrinted(t, "a { & b {} }", "a {\n  & b {\n  }\n}\n", "")
	expectPrinted(t, "a { &:b {} }", "a {\n  &:b {\n  }\n}\n", "")
	expectPrinted(t, "a { &* {} }", "a {\n  &* {\n  }\n}\n", "<stdin>: WARNING: Cannot use type selector \"*\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "a { &|b {} }", "a {\n  &|b {\n  }\n}\n", "<stdin>: WARNING: Cannot use type selector \"|b\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "a { &*|b {} }", "a {\n  &*|b {\n  }\n}\n", "<stdin>: WARNING: Cannot use type selector \"*|b\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "a { &a|b {} }", "a {\n  &a|b {\n  }\n}\n", "<stdin>: WARNING: Cannot use type selector \"a|b\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, "a { &[b] {} }", "a {\n  &[b] {\n  }\n}\n", "")

	expectPrinted(t, "a { && {} }", "a {\n  && {\n  }\n}\n", "")
	expectPrinted(t, "a { & + & {} }", "a {\n  & + & {\n  }\n}\n", "")
	expectPrinted(t, "a { & > & {} }", "a {\n  & > & {\n  }\n}\n", "")
	expectPrinted(t, "a { & ~ & {} }", "a {\n  & ~ & {\n  }\n}\n", "")
	expectPrinted(t, "a { & + c& {} }", "a {\n  & + c& {\n  }\n}\n", "")
	expectPrinted(t, "a { .b& + & {} }", "a {\n  &.b + & {\n  }\n}\n", "")
	expectPrinted(t, "a { .b& + c& {} }", "a {\n  &.b + c& {\n  }\n}\n", "")
	expectPrinted(t, "a { & + & > & ~ & {} }", "a {\n  & + & > & ~ & {\n  }\n}\n", "")

	// CSS nesting works for all tokens except identifiers and functions
	expectPrinted(t, "a { .b {} }", "a {\n  .b {\n  }\n}\n", "")
	expectPrinted(t, "a { #b {} }", "a {\n  #b {\n  }\n}\n", "")
	expectPrinted(t, "a { :b {} }", "a {\n  :b {\n  }\n}\n", "")
	expectPrinted(t, "a { [b] {} }", "a {\n  [b] {\n  }\n}\n", "")
	expectPrinted(t, "a { * {} }", "a {\n  * {\n  }\n}\n", "")
	expectPrinted(t, "a { |b {} }", "a {\n  |b {\n  }\n}\n", "")
	expectPrinted(t, "a { >b {} }", "a {\n  > b {\n  }\n}\n", "")
	expectPrinted(t, "a { +b {} }", "a {\n  + b {\n  }\n}\n", "")
	expectPrinted(t, "a { ~b {} }", "a {\n  ~ b {\n  }\n}\n", "")
	expectPrinted(t, "a { b {} }", "a {\n  b {\n  }\n}\n", "")
	expectPrinted(t, "a { b() {} }", "a {\n  b() {\n  }\n}\n", "<stdin>: WARNING: Unexpected \"b(\"\n")

	// Note: CSS nesting no longer requires each complex selector to contain "&"
	expectPrinted(t, "a { & b, c {} }", "a {\n  & b,\n  c {\n  }\n}\n", "")
	expectPrinted(t, "a { & b, & c {} }", "a {\n  & b,\n  & c {\n  }\n}\n", "")

	// Note: CSS nesting no longer requires the rule to be nested inside a parent
	// (instead un-nested CSS nesting refers to ":scope" or to ":root")
	expectPrinted(t, "& b, c {}", "& b,\nc {\n}\n", "")
	expectPrinted(t, "& b, & c {}", "& b,\n& c {\n}\n", "")
	expectPrinted(t, "b & {}", "b & {\n}\n", "")
	expectPrinted(t, "b &, c {}", "b &,\nc {\n}\n", "")

	expectPrinted(t, "a { .b & { color: red } }", "a {\n  .b & {\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "a { .b& { color: red } }", "a {\n  &.b {\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "a { .b&[c] { color: red } }", "a {\n  &.b[c] {\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "a { &[c] { color: red } }", "a {\n  &[c] {\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "a { [c]& { color: red } }", "a {\n  &[c] {\n    color: red;\n  }\n}\n", "")
	expectPrintedMinify(t, "a { .b & { color: red } }", "a{.b &{color:red}}", "")
	expectPrintedMinify(t, "a { .b& { color: red } }", "a{&.b{color:red}}", "")

	// Nested at-rules
	expectPrinted(t, "a { @media screen { color: red } }", "a {\n  @media screen {\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "a { @media screen { .b { color: green } color: red } }",
		"a {\n  @media screen {\n    .b {\n      color: green;\n    }\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "a { @media screen { color: red; .b { color: green } } }",
		"a {\n  @media screen {\n    color: red;\n    .b {\n      color: green;\n    }\n  }\n}\n", "")
	expectPrinted(t, "html { @layer base { block-size: 100%; @layer support { & body { min-block-size: 100%; } } } }",
		"html {\n  @layer base {\n    block-size: 100%;\n    @layer support {\n      & body {\n        min-block-size: 100%;\n      }\n    }\n  }\n}\n", "")
	expectPrinted(t, ".card { aspect-ratio: 3/4; @scope (&) { :scope { border: 1px solid white } } }",
		".card {\n  aspect-ratio: 3/4;\n  @scope (&) {\n    :scope {\n      border: 1px solid white;\n    }\n  }\n}\n", "")

	// Minify an implicit leading "&"
	expectPrintedMangle(t, "& { color: red }", "& {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "& a { color: red }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "& a, & b { color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "& a, b { color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a, & b { color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "& &a { color: red }", "& &a {\n  color: red;\n}\n", "<stdin>: WARNING: Cannot use type selector \"a\" directly after nesting selector \"&\"\n"+sassWarningMove)
	expectPrintedMangle(t, "& .x { color: red }", ".x {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "& &.x { color: red }", "& &.x {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "& + a { color: red }", "+ a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "& + a& { color: red }", "+ a& {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "&.x { color: red }", "&.x {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a & { color: red }", "a & {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, ".x & { color: red }", ".x & {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "div { & a { color: red } }", "div {\n  & a {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { & .x { color: red } }", "div {\n  .x {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { & .x, & a { color: red } }", "div {\n  .x,\n  a {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { .x, & a { color: red } }", "div {\n  .x,\n  a {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { & a& { color: red } }", "div {\n  & a& {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { & .x { color: red } }", "div {\n  .x {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { & &.x { color: red } }", "div {\n  & &.x {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { & + a { color: red } }", "div {\n  + a {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { & + a& { color: red } }", "div {\n  + a& {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { .x & { color: red } }", "div {\n  .x & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { & div { color: red } }", "@media screen {\n  div {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "a { @media screen { & div { color: red } } }", "a {\n  @media screen {\n    & div {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedMangle(t, "a { :has(& b) { color: red } }", "a {\n  :has(& b) {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "a { :has(& + b) { color: red } }", "a {\n  :has(& + b) {\n    color: red;\n  }\n}\n", "")

	// Reorder selectors to enable removing "&"
	expectPrintedMangle(t, "reorder { & first, .second { color: red } }", "reorder {\n  .second,\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, & .second { color: red } }", "reorder {\n  .second,\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, #second { color: red } }", "reorder {\n  #second,\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, [second] { color: red } }", "reorder {\n  [second],\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, :second { color: red } }", "reorder {\n  :second,\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, + second { color: red } }", "reorder {\n  + second,\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, ~ second { color: red } }", "reorder {\n  ~ second,\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, > second { color: red } }", "reorder {\n  > second,\n  first {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "reorder { & first, second, .third { color: red } }", "reorder {\n  .third,\n  second,\n  first {\n    color: red;\n  }\n}\n", "")

	// Inline no-op nesting
	expectPrintedMangle(t, "div { & { color: red } }", "div {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "div { && { color: red } }", "div {\n  && {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { zoom: 2; & { color: red } }", "div {\n  zoom: 2;\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "div { zoom: 2; && { color: red } }", "div {\n  zoom: 2;\n  && {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div { &, & { color: red } zoom: 2 }", "div {\n  zoom: 2;\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "div { &, && { color: red } zoom: 2 }", "div {\n  &,\n  && {\n    color: red;\n  }\n  zoom: 2;\n}\n", "")
	expectPrintedMangle(t, "div { &&, & { color: red } zoom: 2 }", "div {\n  &&,\n  & {\n    color: red;\n  }\n  zoom: 2;\n}\n", "")
	expectPrintedMangle(t, "div { a: 1; & { b: 4 } b: 2; && { c: 5 } c: 3 }", "div {\n  a: 1;\n  b: 2;\n  && {\n    c: 5;\n  }\n  c: 3;\n  b: 4;\n}\n", "")
	expectPrintedMangle(t, "div { .b { x: 1 } & { x: 2 } }", "div {\n  .b {\n    x: 1;\n  }\n  x: 2;\n}\n", "")
	expectPrintedMangle(t, "div { & { & { & { color: red } } & { & { zoom: 2 } } } }", "div {\n  color: red;\n  zoom: 2;\n}\n", "")

	// Cannot inline no-op nesting with pseudo-elements (https://github.com/w3c/csswg-drafts/issues/7433)
	expectPrintedMangle(t, "div, span:hover { & { color: red } }", "div,\nspan:hover {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "div, span::before { & { color: red } }", "div,\nspan:before {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span:before { & { color: red } }", "div,\nspan:before {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span::after { & { color: red } }", "div,\nspan:after {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span:after { & { color: red } }", "div,\nspan:after {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span::first-line { & { color: red } }", "div,\nspan:first-line {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span:first-line { & { color: red } }", "div,\nspan:first-line {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span::first-letter { & { color: red } }", "div,\nspan:first-letter {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span:first-letter { & { color: red } }", "div,\nspan:first-letter {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span::pseudo { & { color: red } }", "div,\nspan::pseudo {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span:hover { @layer foo { & { color: red } } }", "div,\nspan:hover {\n  @layer foo {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span:hover { @media screen { & { color: red } } }", "div,\nspan:hover {\n  @media screen {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span::pseudo { @layer foo { & { color: red } } }", "div,\nspan::pseudo {\n  @layer foo {\n    & {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedMangle(t, "div, span::pseudo { @media screen { & { color: red } } }", "div,\nspan::pseudo {\n  @media screen {\n    & {\n      color: red;\n    }\n  }\n}\n", "")

	// Lowering tests for nesting
	nestingWarningIs := "<stdin>: WARNING: Transforming this CSS nesting syntax is not supported in the configured target environment\n" +
		"NOTE: The nesting transform for this case must generate an \":is(...)\" but the configured target environment does not support the \":is\" pseudo-class.\n"
	nesting := compat.Nesting
	everything := ^compat.CSSFeature(0)
	expectPrintedLowerUnsupported(t, nesting, ".foo { .bar { color: red } }", ".foo .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { &.bar { color: red } }", ".foo.bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { & .bar { color: red } }", ".foo .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { .baz { color: red } }", ".foo .bar .baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { &.baz { color: red } }", ".foo .bar.baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { & .baz { color: red } }", ".foo .bar .baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { & > .baz { color: red } }", ".foo .bar > .baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { .baz & { color: red } }", ".baz :is(.foo .bar) {\n  color: red;\n}\n", "") // NOT the same as ".baz .foo .bar
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { & .baz & { color: red } }", ".foo .bar .baz :is(.foo .bar) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo, .bar { .baz & { color: red } }", ".baz :is(.foo, .bar) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".foo, .bar { .baz & { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo, [bar~='abc'] { .baz { color: red } }", ":is(.foo, [bar~=abc]) .baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".foo, [bar~='abc'] { .baz { color: red } }", ".foo .baz,\n[bar~=abc] .baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo, [bar~='a b c'] { .baz { color: red } }", ":is(.foo, [bar~=\"a b c\"]) .baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".foo, [bar~='a b c'] { .baz { color: red } }", ".foo .baz,\n[bar~=\"a b c\"] .baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".baz { .foo, .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".baz { .foo, .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".baz { .foo, & .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".baz { .foo, & .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".baz { & .foo, .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".baz { & .foo, .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".baz { & .foo, & .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".baz { & .foo, & .bar { color: red } }", ".baz .foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".baz { .foo, &.bar { color: red } }", ".baz .foo,\n.baz.bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".baz { &.foo, .bar { color: red } }", ".baz.foo,\n.baz .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".baz { &.foo, &.bar { color: red } }", ".baz.foo,\n.baz.bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".baz { &.foo, &.bar { color: red } }", ".baz.foo,\n.baz.bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { color: blue; & .bar { color: red } }", ".foo {\n  color: blue;\n}\n.foo .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { & .bar { color: red } color: blue }", ".foo {\n  color: blue;\n}\n.foo .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { color: blue; & .bar { color: red } zoom: 2 }", ".foo {\n  color: blue;\n  zoom: 2;\n}\n.foo .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a, .b { .c, .d { color: red } }", ":is(.a, .b) .c,\n:is(.a, .b) .d {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".a, .b { .c, .d { color: red } }", ".a .c,\n.a .d,\n.b .c,\n.b .d {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a, .b { & > & { color: red } }", ":is(.a, .b) > :is(.a, .b) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".a, .b { & > & { color: red } }", ".a > .a,\n.a > .b,\n.b > .a,\n.b > .b {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a { color: red; > .b { color: green; > .c { color: blue } } }", ".a {\n  color: red;\n}\n.a > .b {\n  color: green;\n}\n.a > .b > .c {\n  color: blue;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "> .a { color: red; > .b { color: green; > .c { color: blue } } }", "> .a {\n  color: red;\n}\n> .a > .b {\n  color: green;\n}\n> .a > .b > .c {\n  color: blue;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo, .bar, .foo:before, .bar:after { &:hover { color: red } }", ":is(.foo, .bar):hover {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".foo, .bar, .foo:before, .bar:after { &:hover { color: red } }", ".foo:hover,\n.bar:hover {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo, .bar:before { &:hover { color: red } }", ".foo:hover {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo, .bar:before { :hover & { color: red } }", ":hover .foo {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".bar:before { &:hover { color: red } }", ":is():hover {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".bar:before { :hover & { color: red } }", ":hover :is() {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { &:after, & .bar { color: red } }", ".foo:after,\n.foo .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { & .bar, &:after { color: red } }", ".foo .bar,\n.foo:after {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".xy { :where(&.foo) { color: red } }", ":where(.xy.foo) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "div { :where(&.foo) { color: red } }", ":where(div.foo) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".xy { :where(.foo&) { color: red } }", ":where(.xy.foo) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "div { :where(.foo&) { color: red } }", ":where(div.foo) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".xy { :where([href]&) { color: red } }", ":where(.xy[href]) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "div { :where([href]&) { color: red } }", ":where(div[href]) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".xy { :where(:hover&) { color: red } }", ":where(.xy:hover) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "div { :where(:hover&) { color: red } }", ":where(div:hover) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".xy { :where(:is(.foo)&) { color: red } }", ":where(.xy:is(.foo)) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "div { :where(:is(.foo)&) { color: red } }", ":where(div:is(.foo)) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".xy { :where(.foo + &) { color: red } }", ":where(.foo + .xy) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "div { :where(.foo + &) { color: red } }", ":where(.foo + div) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".xy { :where(&, span:is(.foo &)) { color: red } }", ":where(.xy, span:is(.foo .xy)) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "div { :where(&, span:is(.foo &)) { color: red } }", ":where(div, span:is(.foo div)) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "&, a { color: red }", ":scope,\na {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "&, a { .b { color: red } }", ":is(:scope, a) .b {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, "&, a { .b { color: red } }", ":scope .b,\na .b {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "&, a { .b { .c { color: red } } }", ":is(:scope, a) .b .c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, "&, a { .b { .c { color: red } } }", ":scope .b .c,\na .b .c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { > b, > c { color: red } }", "a > b,\na > c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, "a { > b, > c { color: red } }", "a > b,\na > c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { > b, + c { color: red } }", "a > b,\na + c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { & > b, & > c { color: red } }", "a > b,\na > c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, "a { & > b, & > c { color: red } }", "a > b,\na > c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { & > b, & + c { color: red } }", "a > b,\na + c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { > b&, > c& { color: red } }", "a > a:is(b),\na > a:is(c) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, "a { > b&, > c& { color: red } }", "a > a:is(b),\na > a:is(c) {\n  color: red;\n}\n", nestingWarningIs+nestingWarningIs)
	expectPrintedLowerUnsupported(t, nesting, "a { > b&, + c& { color: red } }", "a > a:is(b),\na + a:is(c) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { > &.b, > &.c { color: red } }", "a > a.b,\na > a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, "a { > &.b, > &.c { color: red } }", "a > a.b,\na > a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { > &.b, + &.c { color: red } }", "a > a.b,\na + a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a { > b&, > c& { color: red } }", ".a > b.a,\n.a > c.a {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".a { > b&, > c& { color: red } }", ".a > b.a,\n.a > c.a {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a { > b&, + c& { color: red } }", ".a > b.a,\n.a + c.a {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a { > &.b, > &.c { color: red } }", ".a > .a.b,\n.a > .a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".a { > &.b, > &.c { color: red } }", ".a > .a.b,\n.a > .a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a { > &.b, + &.c { color: red } }", ".a > .a.b,\n.a + .a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "~ .a { > &.b, > &.c { color: red } }", "~ .a > .a.b,\n~ .a > .a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, "~ .a { > &.b, > &.c { color: red } }", "~ .a > .a.b,\n~ .a > .a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "~ .a { > &.b, + &.c { color: red } }", "~ .a > .a.b,\n~ .a + .a.c {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { > &.a, > &.b { color: red } }", ".foo .bar > :is(.foo .bar).a,\n.foo .bar > :is(.foo .bar).b {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, everything, ".foo .bar { > &.a, > &.b { color: red } }", ".foo .bar > :is(.foo .bar).a,\n.foo .bar > :is(.foo .bar).b {\n  color: red;\n}\n", nestingWarningIs+nestingWarningIs)
	expectPrintedLowerUnsupported(t, nesting, ".foo .bar { > &.a, + &.b { color: red } }", ".foo .bar > :is(.foo .bar).a,\n.foo .bar + :is(.foo .bar).b {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".demo { .lg { &.triangle, &.circle { color: red } } }", ".demo .lg.triangle,\n.demo .lg.circle {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".demo { .lg { .triangle, .circle { color: red } } }", ".demo .lg .triangle,\n.demo .lg .circle {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".card { .featured & & & { color: red } }", ".featured .card .card .card {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".a :has(> .c) { .b & { color: red } }", ".b :is(.a :has(> .c)) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { :has(&) { color: red } }", ":has(a) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { :has(> &) { color: red } }", ":has(> a) {\n  color: red;\n}\n", "")

	// Duplicate "&" may be used to increase specificity
	expectPrintedLowerUnsupported(t, nesting, ".foo { &&&.bar { color: red } }", ".foo.foo.foo.bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { &&& .bar { color: red } }", ".foo.foo.foo .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { .bar&&& { color: red } }", ".foo.foo.foo.bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { .bar &&& { color: red } }", ".bar .foo.foo.foo {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { &.bar&.baz& { color: red } }", ".foo.foo.foo.bar.baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { &&&.bar { color: red } }", "a:is(a):is(a).bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { &&& .bar { color: red } }", "a:is(a):is(a) .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { .bar&&& { color: red } }", "a:is(a):is(a).bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { .bar &&& { color: red } }", ".bar a:is(a):is(a) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a { &.bar&.baz& { color: red } }", "a:is(a):is(a).bar.baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a, b { &&&.bar { color: red } }", ":is(a, b):is(a, b):is(a, b).bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a, b { &&& .bar { color: red } }", ":is(a, b):is(a, b):is(a, b) .bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a, b { .bar&&& { color: red } }", ":is(a, b):is(a, b):is(a, b).bar {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a, b { .bar &&& { color: red } }", ".bar :is(a, b):is(a, b):is(a, b) {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, "a, b { &.bar&.baz& { color: red } }", ":is(a, b):is(a, b):is(a, b).bar.baz {\n  color: red;\n}\n", "")
	expectPrintedLowerUnsupported(t, nesting, ".foo { &, &&.bar, &&& .baz { color: red } }", ".foo,\n.foo.foo.bar,\n.foo.foo.foo .baz {\n  color: red;\n}\n", "")

	// These are invalid SASS-style nested suffixes
	expectPrintedLower(t, ".card { &--header { color: red } }", ".card {\n  &--header {\n    color: red;\n  }\n}\n",
		"<stdin>: WARNING: Cannot use type selector \"--header\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrintedLower(t, ".card { &__header { color: red } }", ".card {\n  &__header {\n    color: red;\n  }\n}\n",
		"<stdin>: WARNING: Cannot use type selector \"__header\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrintedLower(t, ".card { .nav &--header { color: red } }", ".card {\n  .nav &--header {\n    color: red;\n  }\n}\n",
		"<stdin>: WARNING: Cannot use type selector \"--header\" directly after nesting selector \"&\"\n"+sassWarningMove)
	expectPrintedLower(t, ".card { .nav &__header { color: red } }", ".card {\n  .nav &__header {\n    color: red;\n  }\n}\n",
		"<stdin>: WARNING: Cannot use type selector \"__header\" directly after nesting selector \"&\"\n"+sassWarningMove)
	expectPrinted(t, ".card { &__header { color: red } }", ".card {\n  &__header {\n    color: red;\n  }\n}\n",
		"<stdin>: WARNING: Cannot use type selector \"__header\" directly after nesting selector \"&\"\n"+sassWarningWrap)
	expectPrinted(t, ".card { .nav &__header { color: red } }", ".card {\n  .nav &__header {\n    color: red;\n  }\n}\n",
		"<stdin>: WARNING: Cannot use type selector \"__header\" directly after nesting selector \"&\"\n"+sassWarningMove)
	expectPrinted(t, ".card { .nav, &__header { color: red } }", ".card {\n  .nav, &__header {\n    color: red;\n  }\n}\n",
		"<stdin>: WARNING: Cannot use type selector \"__header\" directly after nesting selector \"&\"\n"+sassWarningMove)

	// Check pseudo-elements (those coming through "&" must be dropped)
	expectPrintedLower(t, "a, b::before { &.foo { color: red } }", "a.foo {\n  color: red;\n}\n", "")
	expectPrintedLower(t, "a, b::before { & .foo { color: red } }", "a .foo {\n  color: red;\n}\n", "")
	expectPrintedLower(t, "a { &.foo, &::before { color: red } }", "a.foo,\na::before {\n  color: red;\n}\n", "")
	expectPrintedLower(t, "a { & .foo, & ::before { color: red } }", "a .foo,\na ::before {\n  color: red;\n}\n", "")
	expectPrintedLower(t, "a, b::before { color: blue; &.foo, &::after { color: red } }", "a,\nb::before {\n  color: blue;\n}\na.foo,\na::after {\n  color: red;\n}\n", "")

	// Test at-rules
	expectPrintedLower(t, ".foo { @media screen {} }", "", "")
	expectPrintedLower(t, ".foo { @media screen { color: red } }", "@media screen {\n  .foo {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo { @media screen { &:hover { color: red } } }", "@media screen {\n  .foo:hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo { @media screen { :hover { color: red } } }", "@media screen {\n  .foo :hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo, .bar { @media screen { color: red } }", "@media screen {\n  .foo,\n  .bar {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo, .bar { @media screen { &:hover { color: red } } }", "@media screen {\n  .foo:hover,\n  .bar:hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo, .bar { @media screen { :hover { color: red } } }", "@media screen {\n  .foo :hover,\n  .bar :hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo { @layer xyz {} }", "@layer xyz;\n", "")
	expectPrintedLower(t, ".foo { @layer xyz { color: red } }", "@layer xyz {\n  .foo {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo { @layer xyz { &:hover { color: red } } }", "@layer xyz {\n  .foo:hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo { @layer xyz { :hover { color: red } } }", "@layer xyz {\n  .foo :hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo, .bar { @layer xyz { color: red } }", "@layer xyz {\n  .foo,\n  .bar {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo, .bar { @layer xyz { &:hover { color: red } } }", "@layer xyz {\n  .foo:hover,\n  .bar:hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, ".foo, .bar { @layer xyz { :hover { color: red } } }", "@layer xyz {\n  .foo :hover,\n  .bar :hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, "@media screen { @media (min-width: 900px) { a, b { &:hover { color: red } } } }",
		"@media screen {\n  @media (min-width: 900px) {\n    a:hover,\n    b:hover {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedLower(t, "@supports (display: flex) { @supports selector(h2 > p) { a, b { &:hover { color: red } } } }",
		"@supports (display: flex) {\n  @supports selector(h2 > p) {\n    a:hover,\n    b:hover {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedLower(t, "@layer foo { @layer bar { a, b { &:hover { color: red } } } }",
		"@layer foo {\n  @layer bar {\n    a:hover,\n    b:hover {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedLower(t, ".card { @supports (selector(&)) { &:hover { color: red } } }",
		"@supports (selector(&)) {\n  .card:hover {\n    color: red;\n  }\n}\n", "")
	expectPrintedLower(t, "html { @layer base { color: blue; @layer support { & body { color: red } } } }",
		"@layer base {\n  html {\n    color: blue;\n  }\n  @layer support {\n    html body {\n      color: red;\n    }\n  }\n}\n", "")

	// https://github.com/w3c/csswg-drafts/issues/7961#issuecomment-1549874958
	expectPrinted(t, "@media screen { a { x: y } x: y; b { x: y } }", "@media screen {\n  a {\n    x: y;\n  }\n  x: y;\n  b {\n    x: y;\n  }\n}\n", "")
	expectPrinted(t, ":root { @media screen { a { x: y } x: y; b { x: y } } }", ":root {\n  @media screen {\n    a {\n      x: y;\n    }\n    x: y;\n    b {\n      x: y;\n    }\n  }\n}\n", "")
}

func TestBadQualifiedRules(t *testing.T) {
	expectPrinted(t, "$bad: rule;", "$bad: rule; {\n}\n", "<stdin>: WARNING: Unexpected \"$\"\n")
	expectPrinted(t, "$bad: rule; div { color: red }", "$bad: rule; div {\n  color: red;\n}\n", "<stdin>: WARNING: Unexpected \"$\"\n")
	expectPrinted(t, "$bad { color: red }", "$bad {\n  color: red;\n}\n", "<stdin>: WARNING: Unexpected \"$\"\n")
	expectPrinted(t, "a { div.major { color: blue } color: red }", "a {\n  div.major {\n    color: blue;\n  }\n  color: red;\n}\n", "")
	expectPrinted(t, "a { div:hover { color: blue } color: red }", "a {\n  div:hover {\n    color: blue;\n  }\n  color: red;\n}\n", "")
	expectPrinted(t, "a { div:hover { color: blue }; color: red }", "a {\n  div:hover {\n    color: blue;\n  }\n  color: red;\n}\n", "")
	expectPrinted(t, "a { div:hover { color: blue } ; color: red }", "a {\n  div:hover {\n    color: blue;\n  }\n  color: red;\n}\n", "")
	expectPrinted(t, "! { x: y; }", "! {\n  x: y;\n}\n", "<stdin>: WARNING: Unexpected \"!\"\n")
	expectPrinted(t, "! { x: {} }", "! {\n  x: {\n  }\n}\n", "<stdin>: WARNING: Unexpected \"!\"\n<stdin>: WARNING: Expected identifier but found whitespace\n")
	expectPrinted(t, "a { *width: 100%; height: 1px }", "a {\n  *width: 100%;\n  height: 1px;\n}\n", "<stdin>: WARNING: Expected identifier but found \"*\"\n")
	expectPrinted(t, "a { garbage; height: 1px }", "a {\n  garbage;\n  height: 1px;\n}\n", "<stdin>: WARNING: Expected \":\"\n")
	expectPrinted(t, "a { !; height: 1px }", "a {\n  !;\n  height: 1px;\n}\n", "<stdin>: WARNING: Expected identifier but found \"!\"\n")
}

func TestAtRule(t *testing.T) {
	expectPrinted(t, "@unknown", "@unknown;\n", "<stdin>: WARNING: Expected \"{\" but found end of file\n")
	expectPrinted(t, "@unknown;", "@unknown;\n", "")
	expectPrinted(t, "@unknown{}", "@unknown {}\n", "")
	expectPrinted(t, "@unknown x;", "@unknown x;\n", "")
	expectPrinted(t, "@unknown{\na: b;\nc: d;\n}", "@unknown { a: b; c: d; }\n", "")

	expectPrinted(t, "@unknown", "@unknown;\n", "<stdin>: WARNING: Expected \"{\" but found end of file\n")
	expectPrinted(t, "@", "@ {\n}\n", "<stdin>: WARNING: Unexpected \"@\"\n")
	expectPrinted(t, "@;", "@; {\n}\n", "<stdin>: WARNING: Unexpected \"@\"\n")
	expectPrinted(t, "@{}", "@ {\n}\n", "<stdin>: WARNING: Unexpected \"@\"\n")

	expectPrinted(t, "@viewport { width: 100vw }", "@viewport {\n  width: 100vw;\n}\n", "")
	expectPrinted(t, "@-ms-viewport { width: 100vw }", "@-ms-viewport {\n  width: 100vw;\n}\n", "")

	expectPrinted(t, "@document url(\"https://www.example.com/\") { h1 { color: green } }",
		"@document url(https://www.example.com/) {\n  h1 {\n    color: green;\n  }\n}\n", "")
	expectPrinted(t, "@-moz-document url-prefix() { h1 { color: green } }",
		"@-moz-document url-prefix() {\n  h1 {\n    color: green;\n  }\n}\n", "")

	expectPrinted(t, "@media foo { bar }", "@media foo {\n  bar {\n  }\n}\n", "<stdin>: WARNING: Unexpected \"}\"\n")
	expectPrinted(t, "@media foo { bar {}", "@media foo {\n  bar {\n  }\n}\n", "<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "@media foo {", "@media foo {\n}\n", "<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "@media foo", "@media foo {\n}\n", "<stdin>: WARNING: Expected \"{\" but found end of file\n")
	expectPrinted(t, "@media", "@media {\n}\n", "<stdin>: WARNING: Expected \"{\" but found end of file\n")

	// https://www.w3.org/TR/css-page-3/#syntax-page-selector
	expectPrinted(t, `
		@page :first { margin: 0 }
		@page {
			@top-left-corner { content: 'tlc' }
			@top-left { content: 'tl' }
			@top-center { content: 'tc' }
			@top-right { content: 'tr' }
			@top-right-corner { content: 'trc' }
			@bottom-left-corner { content: 'blc' }
			@bottom-left { content: 'bl' }
			@bottom-center { content: 'bc' }
			@bottom-right { content: 'br' }
			@bottom-right-corner { content: 'brc' }
			@left-top { content: 'lt' }
			@left-middle { content: 'lm' }
			@left-bottom { content: 'lb' }
			@right-top { content: 'rt' }
			@right-middle { content: 'rm' }
			@right-bottom { content: 'rb' }
		}
	`, `@page :first {
  margin: 0;
}
@page {
  @top-left-corner {
    content: "tlc";
  }
  @top-left {
    content: "tl";
  }
  @top-center {
    content: "tc";
  }
  @top-right {
    content: "tr";
  }
  @top-right-corner {
    content: "trc";
  }
  @bottom-left-corner {
    content: "blc";
  }
  @bottom-left {
    content: "bl";
  }
  @bottom-center {
    content: "bc";
  }
  @bottom-right {
    content: "br";
  }
  @bottom-right-corner {
    content: "brc";
  }
  @left-top {
    content: "lt";
  }
  @left-middle {
    content: "lm";
  }
  @left-bottom {
    content: "lb";
  }
  @right-top {
    content: "rt";
  }
  @right-middle {
    content: "rm";
  }
  @right-bottom {
    content: "rb";
  }
}
`, "")

	// https://drafts.csswg.org/css-fonts-4/#font-palette-values
	expectPrinted(t, `
		@font-palette-values --Augusta {
			font-family: Handover Sans;
			base-palette: 3;
			override-colors: 1 rgb(43, 12, 9), 2 #000, 3 var(--highlight)
		}
	`, `@font-palette-values --Augusta {
  font-family: Handover Sans;
  base-palette: 3;
  override-colors:
    1 rgb(43, 12, 9),
    2 #000,
    3 var(--highlight);
}
`, "")

	// https://drafts.csswg.org/css-contain-3/#container-rule
	expectPrinted(t, `
		@container my-layout (inline-size > 45em) {
			.foo {
				color: skyblue;
			}
		}
	`, `@container my-layout (inline-size > 45em) {
  .foo {
    color: skyblue;
  }
}
`, "")
	expectPrintedMinify(t, `@container card ( inline-size > 30em ) and style( --responsive = true ) {
	.foo {
		color: skyblue;
	}
}`, "@container card (inline-size > 30em) and style(--responsive = true){.foo{color:skyblue}}", "")
	expectPrintedMangleMinify(t, `@supports ( container-type: size ) {
	@container ( width <= 150px ) {
		#inner {
			background-color: skyblue;
		}
	}
}`, "@supports (container-type: size){@container (width <= 150px){#inner{background-color:#87ceeb}}}", "")

	// https://drafts.csswg.org/css-transitions-2/#defining-before-change-style-the-starting-style-rule
	expectPrinted(t, `
		@starting-style {
			h1 {
				background-color: transparent;
			}
			@layer foo {
				div {
					height: 100px;
				}
			}
		}
	`, `@starting-style {
  h1 {
    background-color: transparent;
  }
  @layer foo {
    div {
      height: 100px;
    }
  }
}
`, "")

	expectPrintedMinify(t, `@starting-style {
	h1 {
		background-color: transparent;
	}
}`, "@starting-style{h1{background-color:transparent}}", "")

	// https://drafts.csswg.org/css-counter-styles/#the-counter-style-rule
	expectPrinted(t, `
		@counter-style box-corner {
			system: fixed;
			symbols: ◰ ◳ ◲ ◱;
			suffix: ': '
		}
	`, `@counter-style box-corner {
  system: fixed;
  symbols: ◰ ◳ ◲ ◱;
  suffix: ": ";
}
`, "")

	// https://drafts.csswg.org/css-fonts/#font-feature-values
	expectPrinted(t, `
		@font-feature-values Roboto {
			font-display: swap;
		}
	`, `@font-feature-values Roboto {
  font-display: swap;
}
`, "")
	expectPrinted(t, `
		@font-feature-values Bongo {
			@swash { ornate: 1 }
		}
	`, `@font-feature-values Bongo {
  @swash {
    ornate: 1;
  }
}
`, "")

	// https://drafts.csswg.org/css-anchor-position-1/#at-ruledef-position-try
	expectPrinted(t, `@position-try --foo { top: 0 }`,
		`@position-try --foo {
  top: 0;
}
`, "")
	expectPrintedMinify(t, `@position-try --foo { top: 0; }`, `@position-try --foo{top:0}`, "")
}

func TestAtCharset(t *testing.T) {
	expectPrinted(t, "@charset \"UTF-8\";", "@charset \"UTF-8\";\n", "")
	expectPrinted(t, "@charset 'UTF-8';", "@charset \"UTF-8\";\n", "")
	expectPrinted(t, "@charset \"utf-8\";", "@charset \"utf-8\";\n", "")
	expectPrinted(t, "@charset \"Utf-8\";", "@charset \"Utf-8\";\n", "")
	expectPrinted(t, "@charset \"US-ASCII\";", "@charset \"US-ASCII\";\n", "<stdin>: WARNING: \"UTF-8\" will be used instead of unsupported charset \"US-ASCII\"\n")
	expectPrinted(t, "@charset;", "@charset;\n", "<stdin>: WARNING: Expected whitespace but found \";\"\n")
	expectPrinted(t, "@charset ;", "@charset;\n", "<stdin>: WARNING: Expected string token but found \";\"\n")
	expectPrinted(t, "@charset\"UTF-8\";", "@charset \"UTF-8\";\n", "<stdin>: WARNING: Expected whitespace but found \"\\\"UTF-8\\\"\"\n")
	expectPrinted(t, "@charset \"UTF-8\"", "@charset \"UTF-8\";\n", "<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrinted(t, "@charset url(UTF-8);", "@charset url(UTF-8);\n", "<stdin>: WARNING: Expected string token but found \"url(UTF-8)\"\n")
	expectPrinted(t, "@charset url(\"UTF-8\");", "@charset url(UTF-8);\n", "<stdin>: WARNING: Expected string token but found \"url(\"\n")
	expectPrinted(t, "@charset \"UTF-8\" ", "@charset \"UTF-8\";\n", "<stdin>: WARNING: Expected \";\" but found whitespace\n")
	expectPrinted(t, "@charset \"UTF-8\"{}", "@charset \"UTF-8\";\n {\n}\n", "<stdin>: WARNING: Expected \";\" but found \"{\"\n")
}

func TestAtImport(t *testing.T) {
	expectPrinted(t, "@import\"foo.css\";", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import \"foo.css\";", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import \"foo.css\" ;", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import url();", "@import \"\";\n", "")
	expectPrinted(t, "@import url(foo.css);", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import url(foo.css) ;", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import url( foo.css );", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import url(\"foo.css\");", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import url(\"foo.css\") ;", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import url( \"foo.css\" );", "@import \"foo.css\";\n", "")
	expectPrinted(t, "@import url(\"foo.css\") print;", "@import \"foo.css\" print;\n", "")
	expectPrinted(t, "@import url(\"foo.css\") screen and (orientation:landscape);", "@import \"foo.css\" screen and (orientation:landscape);\n", "")

	expectPrinted(t, "@import;", "@import;\n", "<stdin>: WARNING: Expected URL token but found \";\"\n")
	expectPrinted(t, "@import ;", "@import;\n", "<stdin>: WARNING: Expected URL token but found \";\"\n")
	expectPrinted(t, "@import \"foo.css\"", "@import \"foo.css\";\n", "<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrinted(t, "@import url(\"foo.css\" extra-stuff);", "@import url(\"foo.css\" extra-stuff);\n", "<stdin>: WARNING: Expected URL token but found \"url(\"\n")
	expectPrinted(t, "@import url(\"foo.css\";", "@import url(\"foo.css\";);\n",
		"<stdin>: WARNING: Expected URL token but found \"url(\"\n<stdin>: WARNING: Expected \")\" to go with \"(\"\n<stdin>: NOTE: The unbalanced \"(\" is here:\n")
	expectPrinted(t, "@import noturl(\"foo.css\");", "@import noturl(\"foo.css\");\n", "<stdin>: WARNING: Expected URL token but found \"noturl(\"\n")
	expectPrinted(t, "@import url(foo.css", "@import \"foo.css\";\n", `<stdin>: WARNING: Expected ")" to end URL token
<stdin>: NOTE: The unbalanced "(" is here:
<stdin>: WARNING: Expected ";" but found end of file
`)

	expectPrinted(t, "@import \"foo.css\" {}", "@import \"foo.css\" {}\n", "<stdin>: WARNING: Expected \";\"\n")
	expectPrinted(t, "@import \"foo\"\na { color: red }\nb { color: blue }", "@import \"foo\" a { color: red }\nb {\n  color: blue;\n}\n", "<stdin>: WARNING: Expected \";\"\n")

	expectPrinted(t, "a { @import \"foo.css\" }", "a {\n  @import \"foo.css\";\n}\n", "<stdin>: WARNING: \"@import\" is only valid at the top level\n<stdin>: WARNING: Expected \";\"\n")
}

func TestLegalComment(t *testing.T) {
	expectPrinted(t, "/*!*/@import \"x\";", "/*!*/\n@import \"x\";\n", "")
	expectPrinted(t, "/*!*/@charset \"UTF-8\";", "/*!*/\n@charset \"UTF-8\";\n", "")
	expectPrinted(t, "/*!*/ @import \"x\";", "/*!*/\n@import \"x\";\n", "")
	expectPrinted(t, "/*!*/ @charset \"UTF-8\";", "/*!*/\n@charset \"UTF-8\";\n", "")
	expectPrinted(t, "/*!*/ @charset \"UTF-8\"; @import \"x\";", "/*!*/\n@charset \"UTF-8\";\n@import \"x\";\n", "")
	expectPrinted(t, "/*!*/ @import \"x\"; @charset \"UTF-8\";", "/*!*/\n@import \"x\";\n@charset \"UTF-8\";\n",
		"<stdin>: WARNING: \"@charset\" must be the first rule in the file\n"+
			"<stdin>: NOTE: This rule cannot come before a \"@charset\" rule\n")

	expectPrinted(t, "@import \"x\";/*!*/", "@import \"x\";\n/*!*/\n", "")
	expectPrinted(t, "@charset \"UTF-8\";/*!*/", "@charset \"UTF-8\";\n/*!*/\n", "")
	expectPrinted(t, "@import \"x\"; /*!*/", "@import \"x\";\n/*!*/\n", "")
	expectPrinted(t, "@charset \"UTF-8\"; /*!*/", "@charset \"UTF-8\";\n/*!*/\n", "")

	expectPrinted(t, "/*! before */ a { --b: var(--c, /*!*/ /*!*/); } /*! after */\n", "/*! before */\na {\n  --b: var(--c, );\n}\n/*! after */\n", "")
}

func TestAtKeyframes(t *testing.T) {
	expectPrinted(t, "@keyframes {}", "@keyframes {}\n", "<stdin>: WARNING: Expected identifier but found \"{\"\n")
	expectPrinted(t, "@keyframes name{}", "@keyframes name {\n}\n", "")
	expectPrinted(t, "@keyframes name {}", "@keyframes name {\n}\n", "")
	expectPrinted(t, "@keyframes name{0%,50%{color:red}25%,75%{color:blue}}",
		"@keyframes name {\n  0%, 50% {\n    color: red;\n  }\n  25%, 75% {\n    color: blue;\n  }\n}\n", "")
	expectPrinted(t, "@keyframes name { 0%, 50% { color: red } 25%, 75% { color: blue } }",
		"@keyframes name {\n  0%, 50% {\n    color: red;\n  }\n  25%, 75% {\n    color: blue;\n  }\n}\n", "")
	expectPrinted(t, "@keyframes name{from{color:red}to{color:blue}}",
		"@keyframes name {\n  from {\n    color: red;\n  }\n  to {\n    color: blue;\n  }\n}\n", "")
	expectPrinted(t, "@keyframes name { from { color: red } to { color: blue } }",
		"@keyframes name {\n  from {\n    color: red;\n  }\n  to {\n    color: blue;\n  }\n}\n", "")

	// Note: Strings as names is allowed in the CSS specification and works in
	// Firefox and Safari but Chrome has strangely decided to deliberately not
	// support this. We always turn all string names into identifiers to avoid
	// them silently breaking in Chrome.
	expectPrinted(t, "@keyframes 'name' {}", "@keyframes name {\n}\n", "")
	expectPrinted(t, "@keyframes 'name 2' {}", "@keyframes name\\ 2 {\n}\n", "")
	expectPrinted(t, "@keyframes 'none' {}", "@keyframes \"none\" {}\n", "")
	expectPrinted(t, "@keyframes 'None' {}", "@keyframes \"None\" {}\n", "")
	expectPrinted(t, "@keyframes 'unset' {}", "@keyframes \"unset\" {}\n", "")
	expectPrinted(t, "@keyframes 'revert' {}", "@keyframes \"revert\" {}\n", "")
	expectPrinted(t, "@keyframes None {}", "@keyframes None {}\n",
		"<stdin>: WARNING: Cannot use \"None\" as a name for \"@keyframes\" without quotes\n"+
			"NOTE: You can put \"None\" in quotes to prevent it from becoming a CSS keyword.\n")
	expectPrinted(t, "@keyframes REVERT {}", "@keyframes REVERT {}\n",
		"<stdin>: WARNING: Cannot use \"REVERT\" as a name for \"@keyframes\" without quotes\n"+
			"NOTE: You can put \"REVERT\" in quotes to prevent it from becoming a CSS keyword.\n")

	expectPrinted(t, "@keyframes name { from { color: red } }", "@keyframes name {\n  from {\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "@keyframes name { 100% { color: red } }", "@keyframes name {\n  100% {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@keyframes name { from { color: red } }", "@keyframes name {\n  0% {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@keyframes name { 100% { color: red } }", "@keyframes name {\n  to {\n    color: red;\n  }\n}\n", "")

	expectPrinted(t, "@-webkit-keyframes name {}", "@-webkit-keyframes name {\n}\n", "")
	expectPrinted(t, "@-moz-keyframes name {}", "@-moz-keyframes name {\n}\n", "")
	expectPrinted(t, "@-ms-keyframes name {}", "@-ms-keyframes name {\n}\n", "")
	expectPrinted(t, "@-o-keyframes name {}", "@-o-keyframes name {\n}\n", "")

	expectPrinted(t, "@keyframes {}", "@keyframes {}\n", "<stdin>: WARNING: Expected identifier but found \"{\"\n")
	expectPrinted(t, "@keyframes name { 0% 100% {} }", "@keyframes name { 0% 100% {} }\n", "<stdin>: WARNING: Expected \",\" but found \"100%\"\n")
	expectPrinted(t, "@keyframes name { {} 0% {} }", "@keyframes name { {} 0% {} }\n", "<stdin>: WARNING: Expected percentage but found \"{\"\n")
	expectPrinted(t, "@keyframes name { 100 {} }", "@keyframes name { 100 {} }\n", "<stdin>: WARNING: Expected percentage but found \"100\"\n")
	expectPrinted(t, "@keyframes name { into {} }", "@keyframes name {\n  into {\n  }\n}\n", "<stdin>: WARNING: Expected percentage but found \"into\"\n")
	expectPrinted(t, "@keyframes name { 1,2 {} }", "@keyframes name { 1, 2 {} }\n", "<stdin>: WARNING: Expected percentage but found \"1\"\n")
	expectPrinted(t, "@keyframes name { 1, 2 {} }", "@keyframes name { 1, 2 {} }\n", "<stdin>: WARNING: Expected percentage but found \"1\"\n")
	expectPrinted(t, "@keyframes name { 1 ,2 {} }", "@keyframes name { 1, 2 {} }\n", "<stdin>: WARNING: Expected percentage but found \"1\"\n")
	expectPrinted(t, "@keyframes name { 1%, {} }", "@keyframes name { 1%, {} }\n", "<stdin>: WARNING: Expected percentage but found \"{\"\n")
	expectPrinted(t, "@keyframes name { 1%, x {} }", "@keyframes name {\n  1%, x {\n  }\n}\n", "<stdin>: WARNING: Expected percentage but found \"x\"\n")
	expectPrinted(t, "@keyframes name { 1%, ! {} }", "@keyframes name { 1%, ! {} }\n", "<stdin>: WARNING: Expected percentage but found \"!\"\n")
	expectPrinted(t, "@keyframes name { .x {} }", "@keyframes name { .x {} }\n", "<stdin>: WARNING: Expected percentage but found \".\"\n")
	expectPrinted(t, "@keyframes name { {} }", "@keyframes name { {} }\n", "<stdin>: WARNING: Expected percentage but found \"{\"\n")
	expectPrinted(t, "@keyframes name { 1% }", "@keyframes name { 1% }\n", "<stdin>: WARNING: Expected \"{\" but found \"}\"\n")
	expectPrinted(t, "@keyframes name { 1%", "@keyframes name { 1% }\n", "<stdin>: WARNING: Expected \"{\" but found end of file\n")
	expectPrinted(t, "@keyframes name { 1%,,2% {} }", "@keyframes name { 1%,, 2% {} }\n", "<stdin>: WARNING: Expected percentage but found \",\"\n")
	expectPrinted(t, "@keyframes name {", "@keyframes name {}\n", "<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")

	expectPrinted(t, "@keyframes x { 1%, {} } @keyframes z { 1% {} }", "@keyframes x { 1%, {} }\n@keyframes z {\n  1% {\n  }\n}\n", "<stdin>: WARNING: Expected percentage but found \"{\"\n")
	expectPrinted(t, "@keyframes x { .y {} } @keyframes z { 1% {} }", "@keyframes x { .y {} }\n@keyframes z {\n  1% {\n  }\n}\n", "<stdin>: WARNING: Expected percentage but found \".\"\n")
	expectPrinted(t, "@keyframes x { x {} } @keyframes z { 1% {} }", "@keyframes x {\n  x {\n  }\n}\n@keyframes z {\n  1% {\n  }\n}\n", "<stdin>: WARNING: Expected percentage but found \"x\"\n")
	expectPrinted(t, "@keyframes x { {} } @keyframes z { 1% {} }", "@keyframes x { {} }\n@keyframes z {\n  1% {\n  }\n}\n", "<stdin>: WARNING: Expected percentage but found \"{\"\n")
	expectPrinted(t, "@keyframes x { 1% {}", "@keyframes x { 1% {} }\n", "<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "@keyframes x { 1% {", "@keyframes x { 1% {} }\n", "<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "@keyframes x { 1%", "@keyframes x { 1% }\n", "<stdin>: WARNING: Expected \"{\" but found end of file\n")
	expectPrinted(t, "@keyframes x {", "@keyframes x {}\n", "<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
}

func TestAnimationName(t *testing.T) {
	// Note: Strings as names is allowed in the CSS specification and works in
	// Firefox and Safari but Chrome has strangely decided to deliberately not
	// support this. We always turn all string names into identifiers to avoid
	// them silently breaking in Chrome.
	expectPrinted(t, "div { animation-name: 'name' }", "div {\n  animation-name: name;\n}\n", "")
	expectPrinted(t, "div { animation-name: 'name 2' }", "div {\n  animation-name: name\\ 2;\n}\n", "")
	expectPrinted(t, "div { animation-name: 'none' }", "div {\n  animation-name: \"none\";\n}\n", "")
	expectPrinted(t, "div { animation-name: 'None' }", "div {\n  animation-name: \"None\";\n}\n", "")
	expectPrinted(t, "div { animation-name: 'unset' }", "div {\n  animation-name: \"unset\";\n}\n", "")
	expectPrinted(t, "div { animation-name: 'revert' }", "div {\n  animation-name: \"revert\";\n}\n", "")
	expectPrinted(t, "div { animation-name: none }", "div {\n  animation-name: none;\n}\n", "")
	expectPrinted(t, "div { animation-name: unset }", "div {\n  animation-name: unset;\n}\n", "")
	expectPrinted(t, "div { animation: 2s linear 'name 2', 3s infinite 'name 3' }", "div {\n  animation: 2s linear name\\ 2, 3s infinite name\\ 3;\n}\n", "")
}

func TestAtRuleValidation(t *testing.T) {
	expectPrinted(t, "a {} b {} c {} @charset \"UTF-8\";", "a {\n}\nb {\n}\nc {\n}\n@charset \"UTF-8\";\n",
		"<stdin>: WARNING: \"@charset\" must be the first rule in the file\n"+
			"<stdin>: NOTE: This rule cannot come before a \"@charset\" rule\n")

	expectPrinted(t, "a {} b {} c {} @import \"foo\";", "a {\n}\nb {\n}\nc {\n}\n@import \"foo\";\n",
		"<stdin>: WARNING: All \"@import\" rules must come first\n"+
			"<stdin>: NOTE: This rule cannot come before an \"@import\" rule\n")
}

func TestAtLayer(t *testing.T) {
	expectPrinted(t, "@layer a, b;", "@layer a, b;\n", "")
	expectPrinted(t, "@layer a {}", "@layer a {\n}\n", "")
	expectPrinted(t, "@layer {}", "@layer {\n}\n", "")
	expectPrinted(t, "@layer a, b {}", "@layer a, b {\n}\n", "<stdin>: WARNING: Expected \";\"\n")
	expectPrinted(t, "@layer;", "@layer;\n", "<stdin>: WARNING: Unexpected \";\"\n")
	expectPrinted(t, "@layer , b {}", "@layer , b {\n}\n", "<stdin>: WARNING: Unexpected \",\"\n")
	expectPrinted(t, "@layer a", "@layer a;\n", "<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrinted(t, "@layer a { @layer b }", "@layer a {\n  @layer b;\n}\n", "<stdin>: WARNING: Expected \";\"\n")
	expectPrinted(t, "@layer a b", "@layer a b;\n", "<stdin>: WARNING: Unexpected \"b\"\n<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrinted(t, "@layer a b ;", "@layer a b;\n", "<stdin>: WARNING: Unexpected \"b\"\n")
	expectPrinted(t, "@layer a b {}", "@layer a b {\n}\n", "<stdin>: WARNING: Unexpected \"b\"\n")

	expectPrinted(t, "@layer a, b;", "@layer a, b;\n", "")
	expectPrinted(t, "@layer a {}", "@layer a {\n}\n", "")
	expectPrinted(t, "@layer {}", "@layer {\n}\n", "")
	expectPrinted(t, "@layer foo { div { color: red } }", "@layer foo {\n  div {\n    color: red;\n  }\n}\n", "")

	// Check semicolon error recovery
	expectPrinted(t, "@layer", "@layer;\n", "<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrinted(t, "@layer a", "@layer a;\n", "<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrinted(t, "@layer a { @layer }", "@layer a {\n  @layer;\n}\n", "<stdin>: WARNING: Expected \";\"\n")
	expectPrinted(t, "@layer a { @layer b }", "@layer a {\n  @layer b;\n}\n", "<stdin>: WARNING: Expected \";\"\n")
	expectPrintedMangle(t, "@layer", "@layer;\n", "<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrintedMangle(t, "@layer a", "@layer a;\n", "<stdin>: WARNING: Expected \";\" but found end of file\n")
	expectPrintedMangle(t, "@layer a { @layer }", "@layer a {\n  @layer;\n}\n", "<stdin>: WARNING: Expected \";\"\n")
	expectPrintedMangle(t, "@layer a { @layer b }", "@layer a.b;\n", "<stdin>: WARNING: Expected \";\"\n")

	// Check mangling
	expectPrintedMangle(t, "@layer foo { div {} }", "@layer foo;\n", "")
	expectPrintedMangle(t, "@layer foo { div { color: yellow } }", "@layer foo {\n  div {\n    color: #ff0;\n  }\n}\n", "")
	expectPrintedMangle(t, "@layer a { @layer b {} }", "@layer a.b;\n", "")
	expectPrintedMangle(t, "@layer a { @layer {} }", "@layer a {\n  @layer {\n  }\n}\n", "")
	expectPrintedMangle(t, "@layer { @layer a {} }", "@layer {\n  @layer a;\n}\n", "")
	expectPrintedMangle(t, "@layer a.b { @layer c.d {} }", "@layer a.b.c.d;\n", "")
	expectPrintedMangle(t, "@layer a.b { @layer c.d {} @layer e.f {} }", "@layer a.b {\n  @layer c.d;\n  @layer e.f;\n}\n", "")
	expectPrintedMangle(t, "@layer a.b { @layer c.d { e { f: g } } }", "@layer a.b.c.d {\n  e {\n    f: g;\n  }\n}\n", "")

	// Invalid layer names should not be merged, since that causes the rule to
	// become invalid. It would be a change in semantics if we merged an invalid
	// rule with a valid rule since then the other valid rule would be invalid.
	initial := "<stdin>: WARNING: \"initial\" cannot be used as a layer name\n"
	inherit := "<stdin>: WARNING: \"inherit\" cannot be used as a layer name\n"
	unset := "<stdin>: WARNING: \"unset\" cannot be used as a layer name\n"
	expectPrinted(t, "@layer foo { @layer initial; }", "@layer foo {\n  @layer initial;\n}\n", initial)
	expectPrinted(t, "@layer foo { @layer inherit; }", "@layer foo {\n  @layer inherit;\n}\n", inherit)
	expectPrinted(t, "@layer foo { @layer unset; }", "@layer foo {\n  @layer unset;\n}\n", unset)
	expectPrinted(t, "@layer initial { @layer foo; }", "@layer initial {\n  @layer foo;\n}\n", initial)
	expectPrinted(t, "@layer inherit { @layer foo; }", "@layer inherit {\n  @layer foo;\n}\n", inherit)
	expectPrinted(t, "@layer unset { @layer foo; }", "@layer unset {\n  @layer foo;\n}\n", unset)
	expectPrintedMangle(t, "@layer foo { @layer initial { a { b: c } } }", "@layer foo {\n  @layer initial {\n    a {\n      b: c;\n    }\n  }\n}\n", initial)
	expectPrintedMangle(t, "@layer initial { @layer foo { a { b: c } } }", "@layer initial {\n  @layer foo {\n    a {\n      b: c;\n    }\n  }\n}\n", initial)

	// Order matters here. Do not drop the first "@layer a;" or the order will be changed.
	expectPrintedMangle(t, "@layer a; @layer b; @layer a;", "@layer a;\n@layer b;\n@layer a;\n", "")

	// Validate ordering with "@layer" and "@import"
	expectPrinted(t, "@layer a; @import url(b);", "@layer a;\n@import \"b\";\n", "")
	expectPrinted(t, "@layer a; @layer b; @import url(c);", "@layer a;\n@layer b;\n@import \"c\";\n", "")
	expectPrinted(t, "@layer a {} @import url(b);", "@layer a {\n}\n@import url(b);\n",
		"<stdin>: WARNING: All \"@import\" rules must come first\n<stdin>: NOTE: This rule cannot come before an \"@import\" rule\n")
	expectPrinted(t, "@import url(a); @layer b; @import url(c);", "@import \"a\";\n@layer b;\n@import url(c);\n",
		"<stdin>: WARNING: All \"@import\" rules must come first\n<stdin>: NOTE: This rule cannot come before an \"@import\" rule\n")
	expectPrinted(t, "@layer a; @charset \"UTF-8\";", "@layer a;\n@charset \"UTF-8\";\n",
		"<stdin>: WARNING: \"@charset\" must be the first rule in the file\n<stdin>: NOTE: This rule cannot come before a \"@charset\" rule\n")
}

func TestEmptyRule(t *testing.T) {
	expectPrinted(t, "div {}", "div {\n}\n", "")
	expectPrinted(t, "@media screen {}", "@media screen {\n}\n", "")
	expectPrinted(t, "@page { @top-left {} }", "@page {\n  @top-left {\n  }\n}\n", "")
	expectPrinted(t, "@keyframes test { from {} to {} }", "@keyframes test {\n  from {\n  }\n  to {\n  }\n}\n", "")

	expectPrintedMangle(t, "div {}", "", "")
	expectPrintedMangle(t, "@media screen {}", "", "")
	expectPrintedMangle(t, "@page { @top-left {} }", "", "")
	expectPrintedMangle(t, "@keyframes test { from {} to {} }", "@keyframes test {\n}\n", "")

	expectPrinted(t, "$invalid {}", "$invalid {\n}\n", "<stdin>: WARNING: Unexpected \"$\"\n")
	expectPrinted(t, "@page { color: red; @top-left {} }", "@page {\n  color: red;\n  @top-left {\n  }\n}\n", "")
	expectPrinted(t, "@keyframes test { from {} to { color: red } }", "@keyframes test {\n  from {\n  }\n  to {\n    color: red;\n  }\n}\n", "")
	expectPrinted(t, "@keyframes test { from { color: red } to {} }", "@keyframes test {\n  from {\n    color: red;\n  }\n  to {\n  }\n}\n", "")

	expectPrintedMangle(t, "$invalid {}", "$invalid {\n}\n", "<stdin>: WARNING: Unexpected \"$\"\n")
	expectPrintedMangle(t, "@page { color: red; @top-left {} }", "@page {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "@keyframes test { from {} to { color: red } }", "@keyframes test {\n  to {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@keyframes test { from { color: red } to {} }", "@keyframes test {\n  0% {\n    color: red;\n  }\n}\n", "")

	expectPrintedMangleMinify(t, "$invalid {}", "$invalid{}", "<stdin>: WARNING: Unexpected \"$\"\n")
	expectPrintedMangleMinify(t, "@page { color: red; @top-left {} }", "@page{color:red}", "")
	expectPrintedMangleMinify(t, "@keyframes test { from {} to { color: red } }", "@keyframes test{to{color:red}}", "")
	expectPrintedMangleMinify(t, "@keyframes test { from { color: red } to {} }", "@keyframes test{0%{color:red}}", "")

	expectPrinted(t, "invalid", "invalid {\n}\n", "<stdin>: WARNING: Expected \"{\" but found end of file\n")
	expectPrinted(t, "invalid }", "invalid } {\n}\n", "<stdin>: WARNING: Unexpected \"}\"\n")
}

func TestMarginAndPaddingAndInset(t *testing.T) {
	for _, x := range []string{"margin", "padding", "inset"} {
		xTop := x + "-top"
		xRight := x + "-right"
		xBottom := x + "-bottom"
		xLeft := x + "-left"
		if x == "inset" {
			xTop = "top"
			xRight = "right"
			xBottom = "bottom"
			xLeft = "left"
		}

		expectPrinted(t, "a { "+x+": 0 1px 0 1px }", "a {\n  "+x+": 0 1px 0 1px;\n}\n", "")
		expectPrinted(t, "a { "+x+": 0 1px 0px 1px }", "a {\n  "+x+": 0 1px 0px 1px;\n}\n", "")

		expectPrintedMangle(t, "a { "+xTop+": 0px }", "a {\n  "+xTop+": 0;\n}\n", "")
		expectPrintedMangle(t, "a { "+xRight+": 0px }", "a {\n  "+xRight+": 0;\n}\n", "")
		expectPrintedMangle(t, "a { "+xBottom+": 0px }", "a {\n  "+xBottom+": 0;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 0px }", "a {\n  "+xLeft+": 0;\n}\n", "")

		expectPrintedMangle(t, "a { "+xTop+": 1px }", "a {\n  "+xTop+": 1px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xRight+": 1px }", "a {\n  "+xRight+": 1px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xBottom+": 1px }", "a {\n  "+xBottom+": 1px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1px }", "a {\n  "+xLeft+": 1px;\n}\n", "")

		expectPrintedMangle(t, "a { "+x+": 0 1px 0 0 }", "a {\n  "+x+": 0 1px 0 0;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1px 2px 1px }", "a {\n  "+x+": 0 1px 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1px 0 1px }", "a {\n  "+x+": 0 1px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 0 0 0 }", "a {\n  "+x+": 0;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 0 0 0 !important }", "a {\n  "+x+": 0 !important;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1px 0px 1px }", "a {\n  "+x+": 0 1px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1 0px 1px }", "a {\n  "+x+": 0 1 0px 1px;\n}\n", "")

		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px; "+xTop+": 5px }", "a {\n  "+x+": 5px 2px 3px 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px; "+xRight+": 5px }", "a {\n  "+x+": 1px 5px 3px 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px; "+xBottom+": 5px }", "a {\n  "+x+": 1px 2px 5px 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px; "+xLeft+": 5px }", "a {\n  "+x+": 1px 2px 3px 5px;\n}\n", "")

		expectPrintedMangle(t, "a { "+xTop+": 5px; "+x+": 1px 2px 3px 4px }", "a {\n  "+x+": 1px 2px 3px 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xRight+": 5px; "+x+": 1px 2px 3px 4px }", "a {\n  "+x+": 1px 2px 3px 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xBottom+": 5px; "+x+": 1px 2px 3px 4px }", "a {\n  "+x+": 1px 2px 3px 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 5px; "+x+": 1px 2px 3px 4px }", "a {\n  "+x+": 1px 2px 3px 4px;\n}\n", "")

		expectPrintedMangle(t, "a { "+xTop+": 1px; "+xTop+": 2px }", "a {\n  "+xTop+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xRight+": 1px; "+xRight+": 2px }", "a {\n  "+xRight+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xBottom+": 1px; "+xBottom+": 2px }", "a {\n  "+xBottom+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1px; "+xLeft+": 2px }", "a {\n  "+xLeft+": 2px;\n}\n", "")

		expectPrintedMangle(t, "a { "+x+": 1px; "+x+": 2px !important }",
			"a {\n  "+x+": 1px;\n  "+x+": 2px !important;\n}\n", "")
		expectPrintedMangle(t, "a { "+xTop+": 1px; "+xTop+": 2px !important }",
			"a {\n  "+xTop+": 1px;\n  "+xTop+": 2px !important;\n}\n", "")
		expectPrintedMangle(t, "a { "+xRight+": 1px; "+xRight+": 2px !important }",
			"a {\n  "+xRight+": 1px;\n  "+xRight+": 2px !important;\n}\n", "")
		expectPrintedMangle(t, "a { "+xBottom+": 1px; "+xBottom+": 2px !important }",
			"a {\n  "+xBottom+": 1px;\n  "+xBottom+": 2px !important;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1px; "+xLeft+": 2px !important }",
			"a {\n  "+xLeft+": 1px;\n  "+xLeft+": 2px !important;\n}\n", "")

		expectPrintedMangle(t, "a { "+x+": 1px !important; "+x+": 2px }",
			"a {\n  "+x+": 1px !important;\n  "+x+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xTop+": 1px !important; "+xTop+": 2px }",
			"a {\n  "+xTop+": 1px !important;\n  "+xTop+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xRight+": 1px !important; "+xRight+": 2px }",
			"a {\n  "+xRight+": 1px !important;\n  "+xRight+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xBottom+": 1px !important; "+xBottom+": 2px }",
			"a {\n  "+xBottom+": 1px !important;\n  "+xBottom+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1px !important; "+xLeft+": 2px }",
			"a {\n  "+xLeft+": 1px !important;\n  "+xLeft+": 2px;\n}\n", "")

		expectPrintedMangle(t, "a { "+xTop+": 1px; "+xTop+": }", "a {\n  "+xTop+": 1px;\n  "+xTop+":;\n}\n", "")
		expectPrintedMangle(t, "a { "+xTop+": 1px; "+xTop+": 2px 3px }", "a {\n  "+xTop+": 1px;\n  "+xTop+": 2px 3px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px; "+xLeft+": -4px; "+xRight+": -2px }", "a {\n  "+x+": 1px -2px 3px -4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px 2px; "+xTop+": 5px }", "a {\n  "+x+": 5px 2px 1px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px; "+xTop+": 5px }", "a {\n  "+x+": 5px 1px 1px;\n}\n", "")

		// This doesn't collapse because if the "calc" has an error it
		// will be ignored and the original rule will show through
		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px; "+xRight+": calc(1px + var(--x)) }",
			"a {\n  "+x+": 1px 2px 3px 4px;\n  "+xRight+": calc(1px + var(--x));\n}\n", "")

		expectPrintedMangle(t, "a { "+xLeft+": 1px; "+xRight+": 2px; "+xTop+": 3px; "+xBottom+": 4px }", "a {\n  "+x+": 3px 2px 4px 1px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px; "+xRight+": 5px !important }",
			"a {\n  "+x+": 1px 2px 3px 4px;\n  "+xRight+": 5px !important;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px 2px 3px 4px !important; "+xRight+": 5px }",
			"a {\n  "+x+": 1px 2px 3px 4px !important;\n  "+xRight+": 5px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1px !important; "+xRight+": 2px; "+xTop+": 3px !important; "+xBottom+": 4px }",
			"a {\n  "+xLeft+": 1px !important;\n  "+xRight+": 2px;\n  "+xTop+": 3px !important;\n  "+xBottom+": 4px;\n}\n", "")

		// This should not be changed because "--x" and "--z" could be empty
		expectPrintedMangle(t, "a { "+x+": var(--x) var(--y) var(--z) var(--y) }", "a {\n  "+x+": var(--x) var(--y) var(--z) var(--y);\n}\n", "")

		// Don't merge different units
		expectPrintedMangle(t, "a { "+x+": 1px; "+x+": 2px; }", "a {\n  "+x+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 1px; "+x+": 2vw; }", "a {\n  "+x+": 1px;\n  "+x+": 2vw;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1px; "+xLeft+": 2px; }", "a {\n  "+xLeft+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1px; "+xLeft+": 2vw; }", "a {\n  "+xLeft+": 1px;\n  "+xLeft+": 2vw;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1px 2cm 3%; "+x+": 4px; }", "a {\n  "+x+": 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1px 2cm 3%; "+x+": 4vw; }", "a {\n  "+x+": 0 1px 2cm 3%;\n  "+x+": 4vw;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1px 2cm 3%; "+xLeft+": 4px; }", "a {\n  "+x+": 0 1px 2cm 4px;\n}\n", "")
		expectPrintedMangle(t, "a { "+x+": 0 1px 2cm 3%; "+xLeft+": 4vw; }", "a {\n  "+x+": 0 1px 2cm 3%;\n  "+xLeft+": 4vw;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1Q; "+xRight+": 2Q; "+xTop+": 3Q; "+xBottom+": 4Q; }", "a {\n  "+x+": 3Q 2Q 4Q 1Q;\n}\n", "")
		expectPrintedMangle(t, "a { "+xLeft+": 1Q; "+xRight+": 2Q; "+xTop+": 3Q; "+xBottom+": 0; }",
			"a {\n  "+xLeft+": 1Q;\n  "+xRight+": 2Q;\n  "+xTop+": 3Q;\n  "+xBottom+": 0;\n}\n", "")
	}

	// "auto" is the only keyword allowed in a quad, and only for "margin" and "inset" not for "padding"
	expectPrintedMangle(t, "a { margin: 1px auto 3px 4px; margin-left: auto }", "a {\n  margin: 1px auto 3px;\n}\n", "")
	expectPrintedMangle(t, "a { inset: 1px auto 3px 4px; left: auto }", "a {\n  inset: 1px auto 3px;\n}\n", "")
	expectPrintedMangle(t, "a { padding: 1px auto 3px 4px; padding-left: auto }", "a {\n  padding: 1px auto 3px 4px;\n  padding-left: auto;\n}\n", "")
	expectPrintedMangle(t, "a { margin: auto; margin-left: 1px }", "a {\n  margin: auto auto auto 1px;\n}\n", "")
	expectPrintedMangle(t, "a { inset: auto; left: 1px }", "a {\n  inset: auto auto auto 1px;\n}\n", "")
	expectPrintedMangle(t, "a { padding: auto; padding-left: 1px }", "a {\n  padding: auto;\n  padding-left: 1px;\n}\n", "")
	expectPrintedMangle(t, "a { margin: inherit; margin-left: 1px }", "a {\n  margin: inherit;\n  margin-left: 1px;\n}\n", "")
	expectPrintedMangle(t, "a { inset: inherit; left: 1px }", "a {\n  inset: inherit;\n  left: 1px;\n}\n", "")
	expectPrintedMangle(t, "a { padding: inherit; padding-left: 1px }", "a {\n  padding: inherit;\n  padding-left: 1px;\n}\n", "")

	expectPrintedLowerMangle(t, "a { top: 0; right: 0; bottom: 0; left: 0; }", "a {\n  top: 0;\n  right: 0;\n  bottom: 0;\n  left: 0;\n}\n", "")

	// "inset" should be expanded when not supported
	expectPrintedLower(t, "a { inset: 0; }", "a {\n  top: 0;\n  right: 0;\n  bottom: 0;\n  left: 0;\n}\n", "")
	expectPrintedLower(t, "a { inset: 0px; }", "a {\n  top: 0px;\n  right: 0px;\n  bottom: 0px;\n  left: 0px;\n}\n", "")
	expectPrintedLower(t, "a { inset: 1px 2px; }", "a {\n  top: 1px;\n  right: 2px;\n  bottom: 1px;\n  left: 2px;\n}\n", "")
	expectPrintedLower(t, "a { inset: 1px 2px 3px; }", "a {\n  top: 1px;\n  right: 2px;\n  bottom: 3px;\n  left: 2px;\n}\n", "")
	expectPrintedLower(t, "a { inset: 1px 2px 3px 4px; }", "a {\n  top: 1px;\n  right: 2px;\n  bottom: 3px;\n  left: 4px;\n}\n", "")

	// When "inset" isn't supported, other mangling should still work
	expectPrintedLowerMangle(t, "a { top: 0px; }", "a {\n  top: 0;\n}\n", "")
	expectPrintedLowerMangle(t, "a { right: 0px; }", "a {\n  right: 0;\n}\n", "")
	expectPrintedLowerMangle(t, "a { bottom: 0px; }", "a {\n  bottom: 0;\n}\n", "")
	expectPrintedLowerMangle(t, "a { left: 0px; }", "a {\n  left: 0;\n}\n", "")
	expectPrintedLowerMangle(t, "a { inset: 0px; }", "a {\n  top: 0;\n  right: 0;\n  bottom: 0;\n  left: 0;\n}\n", "")

	// When "inset" isn't supported, whitespace minifying should still work
	expectPrintedLowerMinify(t, "a { top: 0px; }", "a{top:0px}", "")
	expectPrintedLowerMinify(t, "a { right: 0px; }", "a{right:0px}", "")
	expectPrintedLowerMinify(t, "a { bottom: 0px; }", "a{bottom:0px}", "")
	expectPrintedLowerMinify(t, "a { left: 0px; }", "a{left:0px}", "")
	expectPrintedLowerMinify(t, "a { inset: 0px; }", "a{top:0px;right:0px;bottom:0px;left:0px}", "")
	expectPrintedLowerMinify(t, "a { inset: 1px 2px; }", "a{top:1px;right:2px;bottom:1px;left:2px}", "")
	expectPrintedLowerMinify(t, "a { inset: 1px 2px 3px; }", "a{top:1px;right:2px;bottom:3px;left:2px}", "")
	expectPrintedLowerMinify(t, "a { inset: 1px 2px 3px 4px; }", "a{top:1px;right:2px;bottom:3px;left:4px}", "")
}

func TestBorderRadius(t *testing.T) {
	expectPrinted(t, "a { border-top-left-radius: 0 0 }", "a {\n  border-top-left-radius: 0 0;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 0 0 }", "a {\n  border-top-left-radius: 0;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 0 0px }", "a {\n  border-top-left-radius: 0;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 0 1px }", "a {\n  border-top-left-radius: 0 1px;\n}\n", "")

	expectPrintedMangle(t, "a { border-top-left-radius: 0; border-radius: 1px }", "a {\n  border-radius: 1px;\n}\n", "")

	expectPrintedMangle(t, "a { border-radius: 1px 2px 3px 4px }", "a {\n  border-radius: 1px 2px 3px 4px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1px 2px 1px 3px }", "a {\n  border-radius: 1px 2px 1px 3px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1px 2px 3px 2px }", "a {\n  border-radius: 1px 2px 3px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1px 2px 1px 2px }", "a {\n  border-radius: 1px 2px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1px 1px 1px 1px }", "a {\n  border-radius: 1px;\n}\n", "")

	expectPrintedMangle(t, "a { border-radius: 0/1px 2px 3px 4px }", "a {\n  border-radius: 0 / 1px 2px 3px 4px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0/1px 2px 1px 3px }", "a {\n  border-radius: 0 / 1px 2px 1px 3px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0/1px 2px 3px 2px }", "a {\n  border-radius: 0 / 1px 2px 3px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0/1px 2px 1px 2px }", "a {\n  border-radius: 0 / 1px 2px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0/1px 1px 1px 1px }", "a {\n  border-radius: 0 / 1px;\n}\n", "")

	expectPrintedMangle(t, "a { border-radius: 1px 2px; border-top-left-radius: 3px; }", "a {\n  border-radius: 3px 2px 1px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1px; border-top-left-radius: 3px; }", "a {\n  border-radius: 3px 1px 1px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0/1px; border-top-left-radius: 3px; }", "a {\n  border-radius: 3px 0 0 / 3px 1px 1px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0/1px 2px; border-top-left-radius: 3px; }", "a {\n  border-radius: 3px 0 0 / 3px 2px 1px;\n}\n", "")

	for _, x := range []string{"", "-top-left", "-top-right", "-bottom-left", "-bottom-right"} {
		y := "border" + x + "-radius"
		expectPrintedMangle(t, "a { "+y+": 1px; "+y+": 2px }",
			"a {\n  "+y+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+y+": 1px !important; "+y+": 2px }",
			"a {\n  "+y+": 1px !important;\n  "+y+": 2px;\n}\n", "")
		expectPrintedMangle(t, "a { "+y+": 1px; "+y+": 2px !important }",
			"a {\n  "+y+": 1px;\n  "+y+": 2px !important;\n}\n", "")
		expectPrintedMangle(t, "a { "+y+": 1px !important; "+y+": 2px !important }",
			"a {\n  "+y+": 2px !important;\n}\n", "")

		expectPrintedMangle(t, "a { border-radius: 1px; "+y+": 2px !important; }",
			"a {\n  border-radius: 1px;\n  "+y+": 2px !important;\n}\n", "")
		expectPrintedMangle(t, "a { border-radius: 1px !important; "+y+": 2px; }",
			"a {\n  border-radius: 1px !important;\n  "+y+": 2px;\n}\n", "")
	}

	expectPrintedMangle(t, "a { border-top-left-radius: ; border-radius: 1px }",
		"a {\n  border-top-left-radius:;\n  border-radius: 1px;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 1px; border-radius: / }",
		"a {\n  border-top-left-radius: 1px;\n  border-radius: /;\n}\n", "")

	expectPrintedMangleMinify(t, "a { border-radius: 1px 2px 3px 4px; border-top-right-radius: 5px; }", "a{border-radius:1px 5px 3px 4px}", "")
	expectPrintedMangleMinify(t, "a { border-radius: 1px 2px 3px 4px; border-top-right-radius: 5px 6px; }", "a{border-radius:1px 5px 3px 4px/1px 6px 3px 4px}", "")

	// These should not be changed because "--x" and "--z" could be empty
	expectPrintedMangle(t, "a { border-radius: var(--x) var(--y) var(--z) var(--y) }", "a {\n  border-radius: var(--x) var(--y) var(--z) var(--y);\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0 / var(--x) var(--y) var(--z) var(--y) }", "a {\n  border-radius: 0 / var(--x) var(--y) var(--z) var(--y);\n}\n", "")

	// "inherit" should not be merged
	expectPrintedMangle(t, "a { border-radius: 1px; border-top-left-radius: 0 }", "a {\n  border-radius: 0 1px 1px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: inherit; border-top-left-radius: 0 }", "a {\n  border-radius: inherit;\n  border-top-left-radius: 0;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0; border-top-left-radius: inherit }", "a {\n  border-radius: 0;\n  border-top-left-radius: inherit;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 0; border-radius: inherit }", "a {\n  border-top-left-radius: 0;\n  border-radius: inherit;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: inherit; border-radius: 0 }", "a {\n  border-top-left-radius: inherit;\n  border-radius: 0;\n}\n", "")

	// Don't merge different units
	expectPrintedMangle(t, "a { border-radius: 1px; border-radius: 2px; }", "a {\n  border-radius: 2px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1px; border-top-left-radius: 2px; }", "a {\n  border-radius: 2px 1px 1px;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 1px; border-radius: 2px; }", "a {\n  border-radius: 2px;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 1px; border-top-left-radius: 2px; }", "a {\n  border-top-left-radius: 2px;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1rem; border-radius: 1vw; }", "a {\n  border-radius: 1rem;\n  border-radius: 1vw;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 1rem; border-top-left-radius: 1vw; }", "a {\n  border-radius: 1rem;\n  border-top-left-radius: 1vw;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 1rem; border-radius: 1vw; }", "a {\n  border-top-left-radius: 1rem;\n  border-radius: 1vw;\n}\n", "")
	expectPrintedMangle(t, "a { border-top-left-radius: 1rem; border-top-left-radius: 1vw; }", "a {\n  border-top-left-radius: 1rem;\n  border-top-left-radius: 1vw;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0; border-top-left-radius: 2px; }", "a {\n  border-radius: 2px 0 0;\n}\n", "")
	expectPrintedMangle(t, "a { border-radius: 0; border-top-left-radius: 2rem; }", "a {\n  border-radius: 0;\n  border-top-left-radius: 2rem;\n}\n", "")
}

func TestBoxShadow(t *testing.T) {
	expectPrinted(t, "a { box-shadow: inset 0px 0px 0px 0px black }", "a {\n  box-shadow: inset 0px 0px 0px 0px black;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: 0px 0px 0px 0px inset black }", "a {\n  box-shadow: 0 0 inset #000;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: 0px 0px 0px 0px black inset }", "a {\n  box-shadow: 0 0 #000 inset;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: black 0px 0px 0px 0px inset }", "a {\n  box-shadow: #000 0 0 inset;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: inset 0px 0px 0px 0px black }", "a {\n  box-shadow: inset 0 0 #000;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: inset black 0px 0px 0px 0px }", "a {\n  box-shadow: inset #000 0 0;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: black inset 0px 0px 0px 0px }", "a {\n  box-shadow: #000 inset 0 0;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: yellow 1px 0px 0px 1px inset }", "a {\n  box-shadow: #ff0 1px 0 0 1px inset;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: yellow 1px 0px 1px 0px inset }", "a {\n  box-shadow: #ff0 1px 0 1px inset;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: rebeccapurple, yellow, black }", "a {\n  box-shadow:\n    #639,\n    #ff0,\n    #000;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: 0px 0px 0px var(--foo) black }", "a {\n  box-shadow: 0 0 0 var(--foo) #000;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: 0px 0px 0px 0px var(--foo) black }", "a {\n  box-shadow: 0 0 0 0 var(--foo) #000;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: calc(1px + var(--foo)) 0px 0px 0px black }", "a {\n  box-shadow: calc(1px + var(--foo)) 0 0 0 #000;\n}\n", "")
	expectPrintedMangle(t, "a { box-shadow: inset 0px 0px 0px 0px 0px magenta; }", "a {\n  box-shadow: inset 0 0 0 0 0 #f0f;\n}\n", "")
	expectPrintedMangleMinify(t, "a { box-shadow: rebeccapurple , yellow , black }", "a{box-shadow:#639,#ff0,#000}", "")
	expectPrintedMangleMinify(t, "a { box-shadow: rgb(255, 0, 17) 0 0 1 inset }", "a{box-shadow:#f01 0 0 1 inset}", "")
}

func TestMangleTime(t *testing.T) {
	expectPrintedMangle(t, "a { animation: b 1s }", "a {\n  animation: b 1s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1.s }", "a {\n  animation: b 1s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1.0s }", "a {\n  animation: b 1s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1.02s }", "a {\n  animation: b 1.02s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .1s }", "a {\n  animation: b .1s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .01s }", "a {\n  animation: b .01s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .001s }", "a {\n  animation: b 1ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .0012s }", "a {\n  animation: b 1.2ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b -.001s }", "a {\n  animation: b -1ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b -.0012s }", "a {\n  animation: b -1.2ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .0001s }", "a {\n  animation: b .1ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .00012s }", "a {\n  animation: b .12ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .000123s }", "a {\n  animation: b .123ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .01S }", "a {\n  animation: b .01S;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b .001S }", "a {\n  animation: b 1ms;\n}\n", "")

	expectPrintedMangle(t, "a { animation: b 1ms }", "a {\n  animation: b 1ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 10ms }", "a {\n  animation: b 10ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 100ms }", "a {\n  animation: b .1s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 120ms }", "a {\n  animation: b .12s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 123ms }", "a {\n  animation: b 123ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1000ms }", "a {\n  animation: b 1s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1200ms }", "a {\n  animation: b 1.2s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1230ms }", "a {\n  animation: b 1.23s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1234ms }", "a {\n  animation: b 1234ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b -100ms }", "a {\n  animation: b -.1s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b -120ms }", "a {\n  animation: b -.12s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 120mS }", "a {\n  animation: b .12s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 120Ms }", "a {\n  animation: b .12s;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 123mS }", "a {\n  animation: b 123mS;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 123Ms }", "a {\n  animation: b 123Ms;\n}\n", "")

	// Mangling times with exponents is not currently supported
	expectPrintedMangle(t, "a { animation: b 1e3ms }", "a {\n  animation: b 1e3ms;\n}\n", "")
	expectPrintedMangle(t, "a { animation: b 1E3ms }", "a {\n  animation: b 1E3ms;\n}\n", "")
}

func TestCalc(t *testing.T) {
	expectPrinted(t, "a { b: calc(+(2)) }", "a {\n  b: calc(+(2));\n}\n", "<stdin>: WARNING: \"+\" can only be used as an infix operator, not a prefix operator\n")
	expectPrinted(t, "a { b: calc(-(2)) }", "a {\n  b: calc(-(2));\n}\n", "<stdin>: WARNING: \"-\" can only be used as an infix operator, not a prefix operator\n")
	expectPrinted(t, "a { b: calc(*(2)) }", "a {\n  b: calc(*(2));\n}\n", "")
	expectPrinted(t, "a { b: calc(/(2)) }", "a {\n  b: calc(/(2));\n}\n", "")

	expectPrinted(t, "a { b: calc(1 + 2) }", "a {\n  b: calc(1 + 2);\n}\n", "")
	expectPrinted(t, "a { b: calc(1 - 2) }", "a {\n  b: calc(1 - 2);\n}\n", "")
	expectPrinted(t, "a { b: calc(1 * 2) }", "a {\n  b: calc(1 * 2);\n}\n", "")
	expectPrinted(t, "a { b: calc(1 / 2) }", "a {\n  b: calc(1 / 2);\n}\n", "")

	expectPrinted(t, "a { b: calc(1+ 2) }", "a {\n  b: calc(1+ 2);\n}\n", "<stdin>: WARNING: The \"+\" operator only works if there is whitespace on both sides\n")
	expectPrinted(t, "a { b: calc(1- 2) }", "a {\n  b: calc(1- 2);\n}\n", "<stdin>: WARNING: The \"-\" operator only works if there is whitespace on both sides\n")
	expectPrinted(t, "a { b: calc(1* 2) }", "a {\n  b: calc(1* 2);\n}\n", "")
	expectPrinted(t, "a { b: calc(1/ 2) }", "a {\n  b: calc(1/ 2);\n}\n", "")

	expectPrinted(t, "a { b: calc(1 +2) }", "a {\n  b: calc(1 +2);\n}\n", "<stdin>: WARNING: The \"+\" operator only works if there is whitespace on both sides\n")
	expectPrinted(t, "a { b: calc(1 -2) }", "a {\n  b: calc(1 -2);\n}\n", "<stdin>: WARNING: The \"-\" operator only works if there is whitespace on both sides\n")
	expectPrinted(t, "a { b: calc(1 *2) }", "a {\n  b: calc(1 *2);\n}\n", "")
	expectPrinted(t, "a { b: calc(1 /2) }", "a {\n  b: calc(1 /2);\n}\n", "")

	expectPrinted(t, "a { b: calc(1 +(2)) }", "a {\n  b: calc(1 +(2));\n}\n", "<stdin>: WARNING: The \"+\" operator only works if there is whitespace on both sides\n")
	expectPrinted(t, "a { b: calc(1 -(2)) }", "a {\n  b: calc(1 -(2));\n}\n", "<stdin>: WARNING: The \"-\" operator only works if there is whitespace on both sides\n")
	expectPrinted(t, "a { b: calc(1 *(2)) }", "a {\n  b: calc(1 *(2));\n}\n", "")
	expectPrinted(t, "a { b: calc(1 /(2)) }", "a {\n  b: calc(1 /(2));\n}\n", "")
}

func TestMinifyCalc(t *testing.T) {
	expectPrintedMangleMinify(t, "a { b: calc(x + y) }", "a{b:calc(x + y)}", "")
	expectPrintedMangleMinify(t, "a { b: calc(x - y) }", "a{b:calc(x - y)}", "")
	expectPrintedMangleMinify(t, "a { b: calc(x * y) }", "a{b:calc(x*y)}", "")
	expectPrintedMangleMinify(t, "a { b: calc(x / y) }", "a{b:calc(x/y)}", "")
}

func TestMangleCalc(t *testing.T) {
	expectPrintedMangle(t, "a { b: calc(1) }", "a {\n  b: 1;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc((1)) }", "a {\n  b: 1;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(calc(1)) }", "a {\n  b: 1;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x + y * z) }", "a {\n  b: calc(x + y * z);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x * y + z) }", "a {\n  b: calc(x * y + z);\n}\n", "")

	// Test sum
	expectPrintedMangle(t, "a { b: calc(2 + 3) }", "a {\n  b: 5;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(6 - 2) }", "a {\n  b: 4;\n}\n", "")

	// Test product
	expectPrintedMangle(t, "a { b: calc(2 * 3) }", "a {\n  b: 6;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(6 / 2) }", "a {\n  b: 3;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(2px * 3 + 4px * 5) }", "a {\n  b: 26px;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(2 * 3px + 4 * 5px) }", "a {\n  b: 26px;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(2px * 3 - 4px * 5) }", "a {\n  b: -14px;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(2 * 3px - 4 * 5px) }", "a {\n  b: -14px;\n}\n", "")

	// Test negation
	expectPrintedMangle(t, "a { b: calc(x + 1) }", "a {\n  b: calc(x + 1);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x - 1) }", "a {\n  b: calc(x - 1);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x + -1) }", "a {\n  b: calc(x - 1);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x - -1) }", "a {\n  b: calc(x + 1);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(1 + x) }", "a {\n  b: calc(1 + x);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(1 - x) }", "a {\n  b: calc(1 - x);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(-1 + x) }", "a {\n  b: calc(-1 + x);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(-1 - x) }", "a {\n  b: calc(-1 - x);\n}\n", "")

	// Test inversion
	expectPrintedMangle(t, "a { b: calc(x * 4) }", "a {\n  b: calc(x * 4);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x / 4) }", "a {\n  b: calc(x / 4);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x * 0.25) }", "a {\n  b: calc(x / 4);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(x / 0.25) }", "a {\n  b: calc(x * 4);\n}\n", "")

	// Test operator precedence
	expectPrintedMangle(t, "a { b: calc((a + b) + c) }", "a {\n  b: calc(a + b + c);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(a + (b + c)) }", "a {\n  b: calc(a + b + c);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc((a - b) - c) }", "a {\n  b: calc(a - b - c);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(a - (b - c)) }", "a {\n  b: calc(a - (b - c));\n}\n", "")
	expectPrintedMangle(t, "a { b: calc((a * b) * c) }", "a {\n  b: calc(a * b * c);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(a * (b * c)) }", "a {\n  b: calc(a * b * c);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc((a / b) / c) }", "a {\n  b: calc(a / b / c);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(a / (b / c)) }", "a {\n  b: calc(a / (b / c));\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(a + b * c / d - e) }", "a {\n  b: calc(a + b * c / d - e);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc((a + ((b * c) / d)) - e) }", "a {\n  b: calc(a + b * c / d - e);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc((a + b) * c / (d - e)) }", "a {\n  b: calc((a + b) * c / (d - e));\n}\n", "")

	// Using "var()" should bail because it can expand to any number of tokens
	expectPrintedMangle(t, "a { b: calc(1px - x + 2px) }", "a {\n  b: calc(3px - x);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(1px - var(x) + 2px) }", "a {\n  b: calc(1px - var(x) + 2px);\n}\n", "")

	// Test values that can't be accurately represented as decimals
	expectPrintedMangle(t, "a { b: calc(100% / 1) }", "a {\n  b: 100%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 2) }", "a {\n  b: 50%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 3) }", "a {\n  b: calc(100% / 3);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 4) }", "a {\n  b: 25%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 5) }", "a {\n  b: 20%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 6) }", "a {\n  b: calc(100% / 6);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 7) }", "a {\n  b: calc(100% / 7);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 8) }", "a {\n  b: 12.5%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 9) }", "a {\n  b: calc(100% / 9);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 10) }", "a {\n  b: 10%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 100) }", "a {\n  b: 1%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 1000) }", "a {\n  b: .1%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 10000) }", "a {\n  b: .01%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 100000) }", "a {\n  b: .001%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / 1000000) }", "a {\n  b: calc(100% / 1000000);\n}\n", "") // This actually ends up as "100% * (1 / 1000000)" which is less precise
	expectPrintedMangle(t, "a { b: calc(100% / -1000000) }", "a {\n  b: calc(100% / -1000000);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(100% / -100000) }", "a {\n  b: -.001%;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(3 * (2px + 1em / 7)) }", "a {\n  b: calc(3 * (2px + 1em / 7));\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(3 * (2px + 1em / 8)) }", "a {\n  b: calc(3 * (2px + .125em));\n}\n", "")

	// Non-finite numbers
	expectPrintedMangle(t, "a { b: calc(0px / 0) }", "a {\n  b: calc(0px / 0);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(1px / 0) }", "a {\n  b: calc(1px / 0);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(-1px / 0) }", "a {\n  b: calc(-1px / 0);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(nan) }", "a {\n  b: calc(nan);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(infinity) }", "a {\n  b: calc(infinity);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(-infinity) }", "a {\n  b: calc(-infinity);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(1px / nan) }", "a {\n  b: calc(1px / nan);\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(1px / infinity) }", "a {\n  b: 0px;\n}\n", "")
	expectPrintedMangle(t, "a { b: calc(1px / -infinity) }", "a {\n  b: -0px;\n}\n", "")
}

func TestTransform(t *testing.T) {
	expectPrintedMangle(t, "a { transform: matrix(1, 0, 0, 1, 0, 0) }", "a {\n  transform: scale(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: matrix(2, 0, 0, 1, 0, 0) }", "a {\n  transform: scaleX(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: matrix(1, 0, 0, 2, 0, 0) }", "a {\n  transform: scaleY(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: matrix(2, 0, 0, 3, 0, 0) }", "a {\n  transform: scale(2, 3);\n}\n", "")
	expectPrintedMangle(t, "a { transform: matrix(2, 0, 0, 2, 0, 0) }", "a {\n  transform: scale(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: matrix(1, 0, 0, 1, 1, 2) }",
		"a {\n  transform:\n    matrix(\n      1, 0,\n      0, 1,\n      1, 2);\n}\n", "")

	expectPrintedMangle(t, "a { transform: translate(0, 0) }", "a {\n  transform: translate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(0px, 0px) }", "a {\n  transform: translate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(0%, 0%) }", "a {\n  transform: translate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(1px, 0) }", "a {\n  transform: translate(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(1px, 0px) }", "a {\n  transform: translate(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(1px, 0%) }", "a {\n  transform: translate(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(0, 1px) }", "a {\n  transform: translateY(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(0px, 1px) }", "a {\n  transform: translateY(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(0%, 1px) }", "a {\n  transform: translateY(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(1px, 2px) }", "a {\n  transform: translate(1px, 2px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate(40%, 60%) }", "a {\n  transform: translate(40%, 60%);\n}\n", "")

	expectPrintedMangle(t, "a { transform: translateX(0) }", "a {\n  transform: translate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateX(0px) }", "a {\n  transform: translate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateX(0%) }", "a {\n  transform: translate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateX(1px) }", "a {\n  transform: translate(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateX(50%) }", "a {\n  transform: translate(50%);\n}\n", "")

	expectPrintedMangle(t, "a { transform: translateY(0) }", "a {\n  transform: translateY(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateY(0px) }", "a {\n  transform: translateY(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateY(0%) }", "a {\n  transform: translateY(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateY(1px) }", "a {\n  transform: translateY(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateY(50%) }", "a {\n  transform: translateY(50%);\n}\n", "")

	expectPrintedMangle(t, "a { transform: scale(1) }", "a {\n  transform: scale(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(100%) }", "a {\n  transform: scale(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(10%) }", "a {\n  transform: scale(.1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(99%) }", "a {\n  transform: scale(99%);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(1, 1) }", "a {\n  transform: scale(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(100%, 1) }", "a {\n  transform: scale(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(10%, 0.1) }", "a {\n  transform: scale(.1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(99%, 0.99) }", "a {\n  transform: scale(99%, .99);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(60%, 40%) }", "a {\n  transform: scale(.6, .4);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(3, 1) }", "a {\n  transform: scaleX(3);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(300%, 1) }", "a {\n  transform: scaleX(3);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(1, 3) }", "a {\n  transform: scaleY(3);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale(1, 300%) }", "a {\n  transform: scaleY(3);\n}\n", "")

	expectPrintedMangle(t, "a { transform: scaleX(1) }", "a {\n  transform: scaleX(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleX(2) }", "a {\n  transform: scaleX(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleX(300%) }", "a {\n  transform: scaleX(3);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleX(99%) }", "a {\n  transform: scaleX(99%);\n}\n", "")

	expectPrintedMangle(t, "a { transform: scaleY(1) }", "a {\n  transform: scaleY(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleY(2) }", "a {\n  transform: scaleY(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleY(300%) }", "a {\n  transform: scaleY(3);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleY(99%) }", "a {\n  transform: scaleY(99%);\n}\n", "")

	expectPrintedMangle(t, "a { transform: rotate(0) }", "a {\n  transform: rotate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotate(0deg) }", "a {\n  transform: rotate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotate(1deg) }", "a {\n  transform: rotate(1deg);\n}\n", "")

	expectPrintedMangle(t, "a { transform: skew(0) }", "a {\n  transform: skew(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skew(0deg) }", "a {\n  transform: skew(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skew(1deg) }", "a {\n  transform: skew(1deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skew(1deg, 0) }", "a {\n  transform: skew(1deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skew(1deg, 0deg) }", "a {\n  transform: skew(1deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skew(0, 1deg) }", "a {\n  transform: skew(0, 1deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skew(0deg, 1deg) }", "a {\n  transform: skew(0, 1deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skew(1deg, 2deg) }", "a {\n  transform: skew(1deg, 2deg);\n}\n", "")

	expectPrintedMangle(t, "a { transform: skewX(0) }", "a {\n  transform: skew(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skewX(0deg) }", "a {\n  transform: skew(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skewX(1deg) }", "a {\n  transform: skew(1deg);\n}\n", "")

	expectPrintedMangle(t, "a { transform: skewY(0) }", "a {\n  transform: skewY(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skewY(0deg) }", "a {\n  transform: skewY(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: skewY(1deg) }", "a {\n  transform: skewY(1deg);\n}\n", "")

	expectPrintedMangle(t,
		"a { transform: matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 2) }",
		"a {\n  transform:\n    matrix3d(\n      1, 0, 0, 0,\n      0, 1, 0, 0,\n      0, 0, 1, 0,\n      0, 0, 0, 2);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 2, 3, 4, 1) }",
		"a {\n  transform:\n    matrix3d(\n      1, 0, 0, 0,\n      0, 1, 0, 0,\n      0, 0, 1, 0,\n      2, 3, 4, 1);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1) }",
		"a {\n  transform:\n    matrix3d(\n      1, 0, 1, 0,\n      0, 1, 0, 0,\n      1, 0, 1, 0,\n      0, 0, 0, 1);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1) }",
		"a {\n  transform: scaleZ(1);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(2, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1) }",
		"a {\n  transform: scale3d(2, 1, 1);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1) }",
		"a {\n  transform: scale3d(1, 2, 1);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(2, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1) }",
		"a {\n  transform: scale3d(2, 2, 1);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(2, 0, 0, 0, 0, 3, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1) }",
		"a {\n  transform: scale3d(2, 3, 1);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1) }",
		"a {\n  transform: scaleZ(2);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 0, 1) }",
		"a {\n  transform: scale3d(1, 2, 3);\n}\n", "")
	expectPrintedMangle(t,
		"a { transform: matrix3d(2, 3, 0, 0, 4, 5, 0, 0, 0, 0, 1, 0, 6, 7, 0, 1) }",
		"a {\n  transform:\n    matrix3d(\n      2, 3, 0, 0,\n      4, 5, 0, 0,\n      0, 0, 1, 0,\n      6, 7, 0, 1);\n}\n", "")

	expectPrintedMangle(t, "a { transform: translate3d(0, 0, 0) }", "a {\n  transform: translateZ(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(0%, 0%, 0) }", "a {\n  transform: translateZ(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(0px, 0px, 0px) }", "a {\n  transform: translateZ(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(1px, 0px, 0px) }", "a {\n  transform: translate3d(1px, 0, 0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(0px, 1px, 0px) }", "a {\n  transform: translate3d(0, 1px, 0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(0px, 0px, 1px) }", "a {\n  transform: translateZ(1px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(1px, 2px, 3px) }", "a {\n  transform: translate3d(1px, 2px, 3px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(1px, 0, 3px) }", "a {\n  transform: translate3d(1px, 0, 3px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(0, 2px, 3px) }", "a {\n  transform: translate3d(0, 2px, 3px);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(1px, 2px, 0px) }", "a {\n  transform: translate3d(1px, 2px, 0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translate3d(40%, 60%, 0px) }", "a {\n  transform: translate3d(40%, 60%, 0);\n}\n", "")

	expectPrintedMangle(t, "a { transform: translateZ(0) }", "a {\n  transform: translateZ(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateZ(0px) }", "a {\n  transform: translateZ(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: translateZ(1px) }", "a {\n  transform: translateZ(1px);\n}\n", "")

	expectPrintedMangle(t, "a { transform: scale3d(1, 1, 1) }", "a {\n  transform: scaleZ(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(2, 1, 1) }", "a {\n  transform: scale3d(2, 1, 1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(1, 2, 1) }", "a {\n  transform: scale3d(1, 2, 1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(1, 1, 2) }", "a {\n  transform: scaleZ(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(1, 2, 3) }", "a {\n  transform: scale3d(1, 2, 3);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(2, 3, 1) }", "a {\n  transform: scale3d(2, 3, 1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(2, 2, 1) }", "a {\n  transform: scale3d(2, 2, 1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(3, 300%, 100.00%) }", "a {\n  transform: scale3d(3, 3, 1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scale3d(1%, 2%, 3%) }", "a {\n  transform: scale3d(1%, 2%, 3%);\n}\n", "")

	expectPrintedMangle(t, "a { transform: scaleZ(1) }", "a {\n  transform: scaleZ(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleZ(100%) }", "a {\n  transform: scaleZ(1);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleZ(2) }", "a {\n  transform: scaleZ(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleZ(200%) }", "a {\n  transform: scaleZ(2);\n}\n", "")
	expectPrintedMangle(t, "a { transform: scaleZ(99%) }", "a {\n  transform: scaleZ(99%);\n}\n", "")

	expectPrintedMangle(t, "a { transform: rotate3d(0, 0, 0, 0) }", "a {\n  transform: rotate3d(0, 0, 0, 0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotate3d(0, 0, 0, 0deg) }", "a {\n  transform: rotate3d(0, 0, 0, 0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotate3d(0, 0, 0, 45deg) }", "a {\n  transform: rotate3d(0, 0, 0, 45deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotate3d(1, 0, 0, 45deg) }", "a {\n  transform: rotateX(45deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotate3d(0, 1, 0, 45deg) }", "a {\n  transform: rotateY(45deg);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotate3d(0, 0, 1, 45deg) }", "a {\n  transform: rotate3d(0, 0, 1, 45deg);\n}\n", "")

	expectPrintedMangle(t, "a { transform: rotateX(0) }", "a {\n  transform: rotateX(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotateX(0deg) }", "a {\n  transform: rotateX(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotateX(1deg) }", "a {\n  transform: rotateX(1deg);\n}\n", "")

	expectPrintedMangle(t, "a { transform: rotateY(0) }", "a {\n  transform: rotateY(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotateY(0deg) }", "a {\n  transform: rotateY(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotateY(1deg) }", "a {\n  transform: rotateY(1deg);\n}\n", "")

	expectPrintedMangle(t, "a { transform: rotateZ(0) }", "a {\n  transform: rotate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotateZ(0deg) }", "a {\n  transform: rotate(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: rotateZ(1deg) }", "a {\n  transform: rotate(1deg);\n}\n", "")

	expectPrintedMangle(t, "a { transform: perspective(0) }", "a {\n  transform: perspective(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: perspective(0px) }", "a {\n  transform: perspective(0);\n}\n", "")
	expectPrintedMangle(t, "a { transform: perspective(1px) }", "a {\n  transform: perspective(1px);\n}\n", "")
}

func TestMangleAlpha(t *testing.T) {
	alphas := []string{
		"0", ".004", ".008", ".01", ".016", ".02", ".024", ".027", ".03", ".035", ".04", ".043", ".047", ".05", ".055", ".06",
		".063", ".067", ".07", ".075", ".08", ".082", ".086", ".09", ".094", ".098", ".1", ".106", ".11", ".114", ".118", ".12",
		".125", ".13", ".133", ".137", ".14", ".145", ".15", ".153", ".157", ".16", ".165", ".17", ".173", ".176", ".18", ".184",
		".19", ".192", ".196", ".2", ".204", ".208", ".21", ".216", ".22", ".224", ".227", ".23", ".235", ".24", ".243", ".247",
		".25", ".255", ".26", ".263", ".267", ".27", ".275", ".28", ".282", ".286", ".29", ".294", ".298", ".3", ".306", ".31",
		".314", ".318", ".32", ".325", ".33", ".333", ".337", ".34", ".345", ".35", ".353", ".357", ".36", ".365", ".37", ".373",
		".376", ".38", ".384", ".39", ".392", ".396", ".4", ".404", ".408", ".41", ".416", ".42", ".424", ".427", ".43", ".435",
		".44", ".443", ".447", ".45", ".455", ".46", ".463", ".467", ".47", ".475", ".48", ".482", ".486", ".49", ".494", ".498",
		".5", ".506", ".51", ".514", ".518", ".52", ".525", ".53", ".533", ".537", ".54", ".545", ".55", ".553", ".557", ".56",
		".565", ".57", ".573", ".576", ".58", ".584", ".59", ".592", ".596", ".6", ".604", ".608", ".61", ".616", ".62", ".624",
		".627", ".63", ".635", ".64", ".643", ".647", ".65", ".655", ".66", ".663", ".667", ".67", ".675", ".68", ".682", ".686",
		".69", ".694", ".698", ".7", ".706", ".71", ".714", ".718", ".72", ".725", ".73", ".733", ".737", ".74", ".745", ".75",
		".753", ".757", ".76", ".765", ".77", ".773", ".776", ".78", ".784", ".79", ".792", ".796", ".8", ".804", ".808", ".81",
		".816", ".82", ".824", ".827", ".83", ".835", ".84", ".843", ".847", ".85", ".855", ".86", ".863", ".867", ".87", ".875",
		".88", ".882", ".886", ".89", ".894", ".898", ".9", ".906", ".91", ".914", ".918", ".92", ".925", ".93", ".933", ".937",
		".94", ".945", ".95", ".953", ".957", ".96", ".965", ".97", ".973", ".976", ".98", ".984", ".99", ".992", ".996",
	}

	for i, alpha := range alphas {
		expectPrintedLowerMangle(t, fmt.Sprintf("a { color: #%08X }", i), "a {\n  color: rgba(0, 0, 0, "+alpha+");\n}\n", "")
	}

	// An alpha value of 100% does not use "rgba(...)"
	expectPrintedLowerMangle(t, "a { color: #000000FF }", "a {\n  color: #000;\n}\n", "")
}

func TestMangleDuplicateSelectors(t *testing.T) {
	expectPrinted(t, "a, a { color: red }", "a,\na {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a, a { color: red }", "a {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a, b { color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a, a.foo, a.foo, a.bar, a { color: red }", "a,\na.foo,\na.bar {\n  color: red;\n}\n", "")

	expectPrintedMangle(t, "@media screen { a, a { color: red } }", "@media screen {\n  a {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { a, b { color: red } }", "@media screen {\n  a,\n  b {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { a, a.foo, a.foo, a.bar, a { color: red } }", "@media screen {\n  a,\n  a.foo,\n  a.bar {\n    color: red;\n  }\n}\n", "")
}

func TestMangleDuplicateSelectorRules(t *testing.T) {
	expectPrinted(t, "a { color: red } b { color: red }", "a {\n  color: red;\n}\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } b { color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } div {} b { color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } div { color: red } b { color: red }", "a,\ndiv,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } div { color: red } a { color: red }", "a,\ndiv {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } div { color: blue } b { color: red }", "a {\n  color: red;\n}\ndiv {\n  color: #00f;\n}\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } div { color: blue } a { color: red }", "a {\n  color: red;\n}\ndiv {\n  color: #00f;\n}\na {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red; color: red } b { color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } b { color: red; color: red }", "a,\nb {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } b { color: blue }", "a {\n  color: red;\n}\nb {\n  color: #00f;\n}\n", "")

	// Do not merge duplicates if they are "unsafe"
	expectPrintedMangle(t, "a { color: red } unknown { color: red }", "a {\n  color: red;\n}\nunknown {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "unknown { color: red } a { color: red }", "unknown {\n  color: red;\n}\na {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } video { color: red }", "a {\n  color: red;\n}\nvideo {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "video { color: red } a { color: red }", "video {\n  color: red;\n}\na {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a:last-child { color: red }", "a {\n  color: red;\n}\na:last-child {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a[b=c i] { color: red }", "a {\n  color: red;\n}\na[b=c i] {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } & { color: red }", "a {\n  color: red;\n}\n& {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a + b { color: red }", "a {\n  color: red;\n}\na + b {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a|b { color: red }", "a {\n  color: red;\n}\na|b {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a::hover { color: red }", "a {\n  color: red;\n}\na::hover {\n  color: red;\n}\n", "")

	// Still merge duplicates if they are "safe"
	expectPrintedMangle(t, "a { color: red } a:hover { color: red }", "a,\na:hover {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a[b=c] { color: red }", "a,\na[b=c] {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a#id { color: red }", "a,\na#id {\n  color: red;\n}\n", "")
	expectPrintedMangle(t, "a { color: red } a.cls { color: red }", "a,\na.cls {\n  color: red;\n}\n", "")

	// Skip over comments
	expectPrintedMangle(t, "c { color: green } a { color: red } /*!x*/ /*!y*/ b { color: blue }", "c {\n  color: green;\n}\na {\n  color: red;\n}\n/*!x*/\n/*!y*/\nb {\n  color: #00f;\n}\n", "")
	expectPrintedMangle(t, "c { color: green } a { color: red } /*!x*/ /*!y*/ b { color: red }", "c {\n  color: green;\n}\na,\nb {\n  color: red;\n}\n/*!x*/\n/*!y*/\n", "")
	expectPrintedMangle(t, "c { color: green } a { color: red } /*!x*/ /*!y*/ a { color: red }", "c {\n  color: green;\n}\na {\n  color: red;\n}\n/*!x*/\n/*!y*/\n", "")
}

func TestMangleAtMedia(t *testing.T) {
	expectPrinted(t, "@media screen { @media screen { a { color: red } } }", "@media screen {\n  @media screen {\n    a {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { @media screen { a { color: red } } }", "@media screen {\n  a {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { @media not print { a { color: red } } }", "@media screen {\n  @media not print {\n    a {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { @media not print { @media screen { a { color: red } } } }", "@media screen {\n  @media not print {\n    a {\n      color: red;\n    }\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { a { color: red } @media screen { a { color: red } } }", "@media screen {\n  a {\n    color: red;\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { a { color: red } @media screen { a { color: blue } } }", "@media screen {\n  a {\n    color: red;\n  }\n  a {\n    color: #00f;\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { .a { color: red; @media screen { .b { color: blue } } } }", "@media screen {\n  .a {\n    color: red;\n    .b {\n      color: #00f;\n    }\n  }\n}\n", "")
	expectPrintedMangle(t, "@media screen { a { color: red } } @media screen { b { color: red } }", "@media screen {\n  a {\n    color: red;\n  }\n}\n@media screen {\n  b {\n    color: red;\n  }\n}\n", "")
}

func TestFontWeight(t *testing.T) {
	expectPrintedMangle(t, "a { font-weight: normal }", "a {\n  font-weight: 400;\n}\n", "")
	expectPrintedMangle(t, "a { font-weight: bold }", "a {\n  font-weight: 700;\n}\n", "")
	expectPrintedMangle(t, "a { font-weight: 400 }", "a {\n  font-weight: 400;\n}\n", "")
	expectPrintedMangle(t, "a { font-weight: bolder }", "a {\n  font-weight: bolder;\n}\n", "")
	expectPrintedMangle(t, "a { font-weight: var(--var) }", "a {\n  font-weight: var(--var);\n}\n", "")

	expectPrintedMangleMinify(t, "a { font-weight: normal }", "a{font-weight:400}", "")
}

func TestFontFamily(t *testing.T) {
	expectPrintedMangle(t, "a {font-family: aaa }", "a {\n  font-family: aaa;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: serif }", "a {\n  font-family: serif;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'serif' }", "a {\n  font-family: \"serif\";\n}\n", "")
	expectPrintedMangle(t, "a {font-family: aaa bbb, serif }", "a {\n  font-family: aaa bbb, serif;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa', serif }", "a {\n  font-family: aaa, serif;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: '\"', serif }", "a {\n  font-family: '\"', serif;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa ', serif }", "a {\n  font-family: \"aaa \", serif;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa bbb', serif }", "a {\n  font-family: aaa bbb, serif;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa bbb', 'ccc ddd' }", "a {\n  font-family: aaa bbb, ccc ddd;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa  bbb', serif }", "a {\n  font-family: \"aaa  bbb\", serif;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa serif' }", "a {\n  font-family: \"aaa serif\";\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa bbb', var(--var) }", "a {\n  font-family: \"aaa bbb\", var(--var);\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa bbb', }", "a {\n  font-family: \"aaa bbb\", ;\n}\n", "")
	expectPrintedMangle(t, "a {font-family: , 'aaa bbb' }", "a {\n  font-family: , \"aaa bbb\";\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa',, 'bbb' }", "a {\n  font-family:\n    \"aaa\",,\n    \"bbb\";\n}\n", "")
	expectPrintedMangle(t, "a {font-family: 'aaa bbb', x serif }", "a {\n  font-family: \"aaa bbb\", x serif;\n}\n", "")

	expectPrintedMangleMinify(t, "a {font-family: 'aaa bbb', serif }", "a{font-family:aaa bbb,serif}", "")
	expectPrintedMangleMinify(t, "a {font-family: 'aaa bbb', 'ccc ddd' }", "a{font-family:aaa bbb,ccc ddd}", "")
	expectPrintedMangleMinify(t, "a {font-family: 'initial', serif;}", "a{font-family:\"initial\",serif}", "")
	expectPrintedMangleMinify(t, "a {font-family: 'inherit', serif;}", "a{font-family:\"inherit\",serif}", "")
	expectPrintedMangleMinify(t, "a {font-family: 'unset', serif;}", "a{font-family:\"unset\",serif}", "")
	expectPrintedMangleMinify(t, "a {font-family: 'revert', serif;}", "a{font-family:\"revert\",serif}", "")
	expectPrintedMangleMinify(t, "a {font-family: 'revert-layer', 'Segoe UI', serif;}", "a{font-family:\"revert-layer\",Segoe UI,serif}", "")
	expectPrintedMangleMinify(t, "a {font-family: 'default', serif;}", "a{font-family:\"default\",serif}", "")
}

func TestFont(t *testing.T) {
	expectPrintedMangle(t, "a { font: caption }", "a {\n  font: caption;\n}\n", "")
	expectPrintedMangle(t, "a { font: normal 1px }", "a {\n  font: normal 1px;\n}\n", "")
	expectPrintedMangle(t, "a { font: normal bold }", "a {\n  font: normal bold;\n}\n", "")
	expectPrintedMangle(t, "a { font: 1rem 'aaa bbb' }", "a {\n  font: 1rem aaa bbb;\n}\n", "")
	expectPrintedMangle(t, "a { font: 1rem/1.2 'aaa bbb' }", "a {\n  font: 1rem/1.2 aaa bbb;\n}\n", "")
	expectPrintedMangle(t, "a { font: normal 1rem 'aaa bbb' }", "a {\n  font: 1rem aaa bbb;\n}\n", "")
	expectPrintedMangle(t, "a { font: normal 1rem 'aaa bbb', serif }", "a {\n  font: 1rem aaa bbb, serif;\n}\n", "")
	expectPrintedMangle(t, "a { font: italic small-caps bold ultra-condensed 1rem/1.2 'aaa bbb' }", "a {\n  font: italic small-caps 700 ultra-condensed 1rem/1.2 aaa bbb;\n}\n", "")
	expectPrintedMangle(t, "a { font: oblique 1px 'aaa bbb' }", "a {\n  font: oblique 1px aaa bbb;\n}\n", "")
	expectPrintedMangle(t, "a { font: oblique 45deg 1px 'aaa bbb' }", "a {\n  font: oblique 45deg 1px aaa bbb;\n}\n", "")

	expectPrintedMangle(t, "a { font: var(--var) 'aaa bbb' }", "a {\n  font: var(--var) \"aaa bbb\";\n}\n", "")
	expectPrintedMangle(t, "a { font: normal var(--var) 'aaa bbb' }", "a {\n  font: normal var(--var) \"aaa bbb\";\n}\n", "")
	expectPrintedMangle(t, "a { font: normal 1rem var(--var), 'aaa bbb' }", "a {\n  font: normal 1rem var(--var), \"aaa bbb\";\n}\n", "")

	expectPrintedMangleMinify(t, "a { font: italic small-caps bold ultra-condensed 1rem/1.2 'aaa bbb' }", "a{font:italic small-caps 700 ultra-condensed 1rem/1.2 aaa bbb}", "")
	expectPrintedMangleMinify(t, "a { font: italic small-caps bold ultra-condensed 1rem / 1.2 'aaa bbb' }", "a{font:italic small-caps 700 ultra-condensed 1rem/1.2 aaa bbb}", "")

	// See: https://github.com/evanw/esbuild/issues/3452
	expectPrinted(t, "a { font: 10px'foo' }", "a {\n  font: 10px\"foo\";\n}\n", "")
	expectPrinted(t, "a { font: 10px'123' }", "a {\n  font: 10px\"123\";\n}\n", "")
	expectPrintedMangle(t, "a { font: 10px'foo' }", "a {\n  font: 10px foo;\n}\n", "")
	expectPrintedMangle(t, "a { font: 10px'123' }", "a {\n  font: 10px\"123\";\n}\n", "")
	expectPrintedMangleMinify(t, "a { font: 10px'foo' }", "a{font:10px foo}", "")
	expectPrintedMangleMinify(t, "a { font: 10px'123' }", "a{font:10px\"123\"}", "")
}

func TestWarningUnexpectedCloseBrace(t *testing.T) {
	expectPrinted(t, ".red {\n  color: red;\n}\n}\n.blue {\n  color: blue;\n}\n.green {\n color: green;\n}\n",
		`.red {
  color: red;
}
} .blue {
  color: blue;
}
.green {
  color: green;
}
`,
		`<stdin>: WARNING: Unexpected "}"
`)
}

func TestPropertyTypoWarning(t *testing.T) {
	expectPrinted(t, "a { z-idnex: 0 }", "a {\n  z-idnex: 0;\n}\n", "<stdin>: WARNING: \"z-idnex\" is not a known CSS property\nNOTE: Did you mean \"z-index\" instead?\n")
	expectPrinted(t, "a { x-index: 0 }", "a {\n  x-index: 0;\n}\n", "<stdin>: WARNING: \"x-index\" is not a known CSS property\nNOTE: Did you mean \"z-index\" instead?\n")

	// CSS variables should not be corrected
	expectPrinted(t, "a { --index: 0 }", "a {\n  --index: 0 ;\n}\n", "")

	// Short names should not be corrected ("alt" is actually valid in WebKit, and should not become "all")
	expectPrinted(t, "a { alt: \"\" }", "a {\n  alt: \"\";\n}\n", "")
}

func TestParseErrorRecovery(t *testing.T) {
	expectPrinted(t, "x { y: z", "x {\n  y: z;\n}\n", "<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "x { y: (", "x {\n  y: ();\n}\n", "<stdin>: WARNING: Expected \")\" to go with \"(\"\n<stdin>: NOTE: The unbalanced \"(\" is here:\n")
	expectPrinted(t, "x { y: [", "x {\n  y: [];\n}\n", "<stdin>: WARNING: Expected \"]\" to go with \"[\"\n<stdin>: NOTE: The unbalanced \"[\" is here:\n")
	expectPrinted(t, "x { y: {", "x {\n  y: {\n  }\n}\n",
		"<stdin>: WARNING: Expected identifier but found whitespace\n<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "x { y: z(", "x {\n  y: z();\n}\n", "<stdin>: WARNING: Expected \")\" to go with \"(\"\n<stdin>: NOTE: The unbalanced \"(\" is here:\n")
	expectPrinted(t, "x { y: z(abc", "x {\n  y: z(abc);\n}\n", "<stdin>: WARNING: Expected \")\" to go with \"(\"\n<stdin>: NOTE: The unbalanced \"(\" is here:\n")
	expectPrinted(t, "x { y: url(", "x {\n  y: url();\n}\n",
		"<stdin>: WARNING: Expected \")\" to end URL token\n<stdin>: NOTE: The unbalanced \"(\" is here:\n<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "x { y: url(abc", "x {\n  y: url(abc);\n}\n",
		"<stdin>: WARNING: Expected \")\" to end URL token\n<stdin>: NOTE: The unbalanced \"(\" is here:\n<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "x { y: url(; }", "x {\n  y: url(; };\n}\n",
		"<stdin>: WARNING: Expected \")\" to end URL token\n<stdin>: NOTE: The unbalanced \"(\" is here:\n<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "x { y: url(abc;", "x {\n  y: url(abc;);\n}\n",
		"<stdin>: WARNING: Expected \")\" to end URL token\n<stdin>: NOTE: The unbalanced \"(\" is here:\n<stdin>: WARNING: Expected \"}\" to go with \"{\"\n<stdin>: NOTE: The unbalanced \"{\" is here:\n")
	expectPrinted(t, "/* @license */ x {} /* @preserve", "/* @license */\nx {\n}\n",
		"<stdin>: ERROR: Expected \"*/\" to terminate multi-line comment\n<stdin>: NOTE: The multi-line comment starts here:\n")
	expectPrinted(t, "a { b: c; d: 'e\n f: g; h: i }", "a {\n  b: c;\n  d: 'e\n  f: g;\n  h: i;\n}\n", "<stdin>: WARNING: Unterminated string token\n")
	expectPrintedMinify(t, "a { b: c; d: 'e\n f: g; h: i }", "a{b:c;d:'e\nf: g;h:i}", "<stdin>: WARNING: Unterminated string token\n")
}

func TestPrefixInsertion(t *testing.T) {
	// General "-webkit-" tests
	for _, key := range []string{
		"backdrop-filter",
		"box-decoration-break",
		"clip-path",
		"font-kerning",
		"initial-letter",
		"mask-image",
		"mask-origin",
		"mask-position",
		"mask-repeat",
		"mask-size",
		"print-color-adjust",
		"text-decoration-skip",
		"text-emphasis-color",
		"text-emphasis-position",
		"text-emphasis-style",
		"text-orientation",
	} {
		expectPrintedWithAllPrefixes(t,
			"a { "+key+": url(x.png) }",
			"a {\n  -webkit-"+key+": url(x.png);\n  "+key+": url(x.png);\n}\n", "")

		expectPrintedWithAllPrefixes(t,
			"a { before: value; "+key+": url(x.png) }",
			"a {\n  before: value;\n  -webkit-"+key+": url(x.png);\n  "+key+": url(x.png);\n}\n", "")

		expectPrintedWithAllPrefixes(t,
			"a { "+key+": url(x.png); after: value }",
			"a {\n  -webkit-"+key+": url(x.png);\n  "+key+": url(x.png);\n  after: value;\n}\n", "")

		expectPrintedWithAllPrefixes(t,
			"a { before: value; "+key+": url(x.png); after: value }",
			"a {\n  before: value;\n  -webkit-"+key+": url(x.png);\n  "+key+": url(x.png);\n  after: value;\n}\n", "")

		expectPrintedWithAllPrefixes(t,
			"a {\n  -webkit-"+key+": url(x.png);\n  "+key+": url(y.png);\n}\n",
			"a {\n  -webkit-"+key+": url(x.png);\n  "+key+": url(y.png);\n}\n", "")

		expectPrintedWithAllPrefixes(t,
			"a {\n  "+key+": url(y.png);\n  -webkit-"+key+": url(x.png);\n}\n",
			"a {\n  "+key+": url(y.png);\n  -webkit-"+key+": url(x.png);\n}\n", "")

		expectPrintedWithAllPrefixes(t,
			"a { "+key+": url(x.png); "+key+": url(y.png) }",
			"a {\n  -webkit-"+key+": url(x.png);\n  "+key+": url(x.png);\n  -webkit-"+key+": url(y.png);\n  "+key+": url(y.png);\n}\n", "")
	}

	// Special-case tests
	expectPrintedWithAllPrefixes(t, "a { appearance: none }", "a {\n  -webkit-appearance: none;\n  -moz-appearance: none;\n  appearance: none;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { background-clip: not-text }", "a {\n  background-clip: not-text;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { background-clip: text !important }",
		"a {\n  -webkit-background-clip: text !important;\n  -ms-background-clip: text !important;\n  background-clip: text !important;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { background-clip: text }", "a {\n  -webkit-background-clip: text;\n  -ms-background-clip: text;\n  background-clip: text;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { hyphens: auto }", "a {\n  -webkit-hyphens: auto;\n  -moz-hyphens: auto;\n  -ms-hyphens: auto;\n  hyphens: auto;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { position: absolute }", "a {\n  position: absolute;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { position: sticky !important }", "a {\n  position: -webkit-sticky !important;\n  position: sticky !important;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { position: sticky }", "a {\n  position: -webkit-sticky;\n  position: sticky;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { tab-size: 2 }", "a {\n  -moz-tab-size: 2;\n  -o-tab-size: 2;\n  tab-size: 2;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { text-decoration-color: none }", "a {\n  -webkit-text-decoration-color: none;\n  -moz-text-decoration-color: none;\n  text-decoration-color: none;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { text-decoration-line: none }", "a {\n  -webkit-text-decoration-line: none;\n  -moz-text-decoration-line: none;\n  text-decoration-line: none;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { text-size-adjust: none }", "a {\n  -webkit-text-size-adjust: none;\n  -ms-text-size-adjust: none;\n  text-size-adjust: none;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { user-select: none }",
		"a {\n  -webkit-user-select: none;\n  -khtml-user-select: none;\n  -moz-user-select: -moz-none;\n  -ms-user-select: none;\n  user-select: none;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { mask-composite: add, subtract, intersect, exclude }",
		"a {\n  -webkit-mask-composite:\n    source-over,\n    source-out,\n    source-in,\n    xor;\n  mask-composite:\n    add,\n    subtract,\n    intersect,\n    exclude;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { width: stretch }",
		"a {\n  width: -webkit-fill-available;\n  width: -moz-available;\n  width: stretch;\n}\n", "")

	// Check that we insert prefixed rules each time an unprefixed rule is
	// encountered. This matches the behavior of the popular "autoprefixer" tool.
	expectPrintedWithAllPrefixes(t,
		"a { before: value; mask-image: a; middle: value; mask-image: b; after: value }",
		"a {\n  before: value;\n  -webkit-mask-image: a;\n  mask-image: a;\n  middle: value;\n  -webkit-mask-image: b;\n  mask-image: b;\n  after: value;\n}\n", "")

	// Test that we don't insert duplicated rules when source code is processed
	// twice. This matches the behavior of the popular "autoprefixer" tool.
	expectPrintedWithAllPrefixes(t,
		"a { before: value; -webkit-text-size-adjust: 1; -ms-text-size-adjust: 2; text-size-adjust: 3; after: value }",
		"a {\n  before: value;\n  -webkit-text-size-adjust: 1;\n  -ms-text-size-adjust: 2;\n  text-size-adjust: 3;\n  after: value;\n}\n", "")
	expectPrintedWithAllPrefixes(t,
		"a { before: value; -webkit-text-size-adjust: 1; text-size-adjust: 3; after: value }",
		"a {\n  before: value;\n  -webkit-text-size-adjust: 1;\n  -ms-text-size-adjust: 3;\n  text-size-adjust: 3;\n  after: value;\n}\n", "")
	expectPrintedWithAllPrefixes(t,
		"a { before: value; -ms-text-size-adjust: 2; text-size-adjust: 3; after: value }",
		"a {\n  before: value;\n  -ms-text-size-adjust: 2;\n  -webkit-text-size-adjust: 3;\n  text-size-adjust: 3;\n  after: value;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { width: -moz-available; width: stretch }",
		"a {\n  width: -moz-available;\n  width: -webkit-fill-available;\n  width: stretch;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { width: -webkit-fill-available; width: stretch }",
		"a {\n  width: -webkit-fill-available;\n  width: -moz-available;\n  width: stretch;\n}\n", "")
	expectPrintedWithAllPrefixes(t, "a { width: -webkit-fill-available; width: -moz-available; width: stretch }",
		"a {\n  width: -webkit-fill-available;\n  width: -moz-available;\n  width: stretch;\n}\n", "")
}

func TestNthChild(t *testing.T) {
	for _, nth := range []string{"nth-child", "nth-last-child"} {
		expectPrinted(t, ":"+nth+"(x) {}", ":"+nth+"(x) {\n}\n", "<stdin>: WARNING: Unexpected \"x\"\n")
		expectPrinted(t, ":"+nth+"(1e2) {}", ":"+nth+"(1e2) {\n}\n", "<stdin>: WARNING: Unexpected \"1e2\"\n")
		expectPrinted(t, ":"+nth+"(-n-) {}", ":"+nth+"(-n-) {\n}\n", "<stdin>: WARNING: Expected number but found \")\"\n")
		expectPrinted(t, ":"+nth+"(-nn) {}", ":"+nth+"(-nn) {\n}\n", "<stdin>: WARNING: Unexpected \"-nn\"\n")
		expectPrinted(t, ":"+nth+"(-n-n) {}", ":"+nth+"(-n-n) {\n}\n", "<stdin>: WARNING: Unexpected \"-n-n\"\n")
		expectPrinted(t, ":"+nth+"(-2n-) {}", ":"+nth+"(-2n-) {\n}\n", "<stdin>: WARNING: Expected number but found \")\"\n")
		expectPrinted(t, ":"+nth+"(-2n-2n) {}", ":"+nth+"(-2n-2n) {\n}\n", "<stdin>: WARNING: Unexpected \"-2n-2n\"\n")
		expectPrinted(t, ":"+nth+"(+) {}", ":"+nth+"(+) {\n}\n", "<stdin>: WARNING: Unexpected \")\"\n")
		expectPrinted(t, ":"+nth+"(-) {}", ":"+nth+"(-) {\n}\n", "<stdin>: WARNING: Unexpected \"-\"\n")
		expectPrinted(t, ":"+nth+"(+ 2) {}", ":"+nth+"(+ 2) {\n}\n", "<stdin>: WARNING: Unexpected whitespace\n")
		expectPrinted(t, ":"+nth+"(- 2) {}", ":"+nth+"(- 2) {\n}\n", "<stdin>: WARNING: Unexpected \"-\"\n")

		expectPrinted(t, ":"+nth+"(0) {}", ":"+nth+"(0) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(0 ) {}", ":"+nth+"(0) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( 0) {}", ":"+nth+"(0) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(00) {}", ":"+nth+"(0) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(01) {}", ":"+nth+"(1) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(0n) {}", ":"+nth+"(0n) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(n) {}", ":"+nth+"(n) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(-n) {}", ":"+nth+"(-n) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(1n) {}", ":"+nth+"(n) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(-1n) {}", ":"+nth+"(-n) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n) {}", ":"+nth+"(2n) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(-2n) {}", ":"+nth+"(-2n) {\n}\n", "")

		expectPrinted(t, ":"+nth+"(odd) {}", ":"+nth+"(odd) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(odd ) {}", ":"+nth+"(odd) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( odd) {}", ":"+nth+"(odd) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(even) {}", ":"+nth+"(even) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(even ) {}", ":"+nth+"(even) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( even) {}", ":"+nth+"(even) {\n}\n", "")

		expectPrinted(t, ":"+nth+"(n+3) {}", ":"+nth+"(n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(n-3) {}", ":"+nth+"(n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(n +3) {}", ":"+nth+"(n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(n -3) {}", ":"+nth+"(n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(n+ 3) {}", ":"+nth+"(n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(n- 3) {}", ":"+nth+"(n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( n + 3 ) {}", ":"+nth+"(n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( n - 3 ) {}", ":"+nth+"(n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(+n+3) {}", ":"+nth+"(n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(+n-3) {}", ":"+nth+"(n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(-n+3) {}", ":"+nth+"(-n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(-n-3) {}", ":"+nth+"(-n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( +n + 3 ) {}", ":"+nth+"(n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( +n - 3 ) {}", ":"+nth+"(n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( -n + 3 ) {}", ":"+nth+"(-n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( -n - 3 ) {}", ":"+nth+"(-n-3) {\n}\n", "")

		expectPrinted(t, ":"+nth+"(2n+3) {}", ":"+nth+"(2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n-3) {}", ":"+nth+"(2n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n +3) {}", ":"+nth+"(2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n -3) {}", ":"+nth+"(2n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n+ 3) {}", ":"+nth+"(2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n- 3) {}", ":"+nth+"(2n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( 2n + 3 ) {}", ":"+nth+"(2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( 2n - 3 ) {}", ":"+nth+"(2n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(+2n+3) {}", ":"+nth+"(2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(+2n-3) {}", ":"+nth+"(2n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(-2n+3) {}", ":"+nth+"(-2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(-2n-3) {}", ":"+nth+"(-2n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( +2n + 3 ) {}", ":"+nth+"(2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( +2n - 3 ) {}", ":"+nth+"(2n-3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( -2n + 3 ) {}", ":"+nth+"(-2n+3) {\n}\n", "")
		expectPrinted(t, ":"+nth+"( -2n - 3 ) {}", ":"+nth+"(-2n-3) {\n}\n", "")

		expectPrinted(t, ":"+nth+"(2n of + .foo) {}", ":"+nth+"(2n of + .foo) {\n}\n", "<stdin>: WARNING: Unexpected \"+\"\n")
		expectPrinted(t, ":"+nth+"(2n of .foo, ~.bar) {}", ":"+nth+"(2n of .foo, ~.bar) {\n}\n", "<stdin>: WARNING: Unexpected \"~\"\n")
		expectPrinted(t, ":"+nth+"(2n of .foo) {}", ":"+nth+"(2n of .foo) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n of.foo+.bar) {}", ":"+nth+"(2n of .foo + .bar) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n of[href]) {}", ":"+nth+"(2n of [href]) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n of.foo,.bar) {}", ":"+nth+"(2n of .foo, .bar) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n of .foo, .bar) {}", ":"+nth+"(2n of .foo, .bar) {\n}\n", "")
		expectPrinted(t, ":"+nth+"(2n of .foo , .bar ) {}", ":"+nth+"(2n of .foo, .bar) {\n}\n", "")

		expectPrintedMinify(t, ":"+nth+"(2n of [foo] , [bar] ) {}", ":"+nth+"(2n of[foo],[bar]){}", "")
		expectPrintedMinify(t, ":"+nth+"(2n of .foo , .bar ) {}", ":"+nth+"(2n of.foo,.bar){}", "")
		expectPrintedMinify(t, ":"+nth+"(2n of #foo , #bar ) {}", ":"+nth+"(2n of#foo,#bar){}", "")
		expectPrintedMinify(t, ":"+nth+"(2n of :foo , :bar ) {}", ":"+nth+"(2n of:foo,:bar){}", "")
		expectPrintedMinify(t, ":"+nth+"(2n of div , span ) {}", ":"+nth+"(2n of div,span){}", "")

		expectPrintedMangle(t, ":"+nth+"(even) { color: red }", ":"+nth+"(2n) {\n  color: red;\n}\n", "")
		expectPrintedMangle(t, ":"+nth+"(2n+1) { color: red }", ":"+nth+"(odd) {\n  color: red;\n}\n", "")
		expectPrintedMangle(t, ":"+nth+"(0n) { color: red }", ":"+nth+"(0) {\n  color: red;\n}\n", "")
		expectPrintedMangle(t, ":"+nth+"(0n+0) { color: red }", ":"+nth+"(0) {\n  color: red;\n}\n", "")
		expectPrintedMangle(t, ":"+nth+"(1n+0) { color: red }", ":"+nth+"(n) {\n  color: red;\n}\n", "")
		expectPrintedMangle(t, ":"+nth+"(0n-2) { color: red }", ":"+nth+"(-2) {\n  color: red;\n}\n", "")
		expectPrintedMangle(t, ":"+nth+"(0n+2) { color: red }", ":"+nth+"(2) {\n  color: red;\n}\n", "")
	}

	for _, nth := range []string{"nth-of-type", "nth-last-of-type"} {
		expectPrinted(t, ":"+nth+"(2n of .foo) {}", ":"+nth+"(2n of .foo) {\n}\n",
			"<stdin>: WARNING: Expected \")\" to go with \"(\"\n<stdin>: NOTE: The unbalanced \"(\" is here:\n")
		expectPrinted(t, ":"+nth+"(+2n + 1) {}", ":"+nth+"(2n+1) {\n}\n", "")
	}
}

func TestComposes(t *testing.T) {
	expectPrinted(t, ".foo { composes: bar; color: red }", ".foo {\n  composes: bar;\n  color: red;\n}\n", "")
	expectPrinted(t, ".foo .bar { composes: bar; color: red }", ".foo .bar {\n  composes: bar;\n  color: red;\n}\n", "")
	expectPrinted(t, ".foo, .bar { composes: bar; color: red }", ".foo,\n.bar {\n  composes: bar;\n  color: red;\n}\n", "")
	expectPrintedLocal(t, ".foo { composes: bar; color: red }", ".foo {\n  color: red;\n}\n", "")
	expectPrintedLocal(t, ".foo { composes: bar baz; color: red }", ".foo {\n  color: red;\n}\n", "")
	expectPrintedLocal(t, ".foo { composes: bar from global; color: red }", ".foo {\n  color: red;\n}\n", "")
	expectPrintedLocal(t, ".foo { composes: bar from \"file.css\"; color: red }", ".foo {\n  color: red;\n}\n", "")
	expectPrintedLocal(t, ".foo { composes: bar from url(file.css); color: red }", ".foo {\n  color: red;\n}\n", "")
	expectPrintedLocal(t, ".foo { & { composes: bar; color: red } }", ".foo {\n  & {\n    color: red;\n  }\n}\n", "")
	expectPrintedLocal(t, ".foo { :local { composes: bar; color: red } }", ".foo {\n  color: red;\n}\n", "")
	expectPrintedLocal(t, ".foo { :global { composes: bar; color: red } }", ".foo {\n  color: red;\n}\n", "")

	expectPrinted(t, ".foo, .bar { composes: bar from github }", ".foo,\n.bar {\n  composes: bar from github;\n}\n", "")
	expectPrintedLocal(t, ".foo { composes: bar from github }", ".foo {\n}\n", "<stdin>: WARNING: \"composes\" declaration uses invalid location \"github\"\n")

	badComposes := "<stdin>: WARNING: \"composes\" only works inside single class selectors\n" +
		"<stdin>: NOTE: The parent selector is not a single class selector because of the syntax here:\n"
	expectPrintedLocal(t, "& { composes: bar; color: red }", "& {\n  color: red;\n}\n", badComposes)
	expectPrintedLocal(t, ".foo& { composes: bar; color: red }", "&.foo {\n  color: red;\n}\n", badComposes)
	expectPrintedLocal(t, ".foo.bar { composes: bar; color: red }", ".foo.bar {\n  color: red;\n}\n", badComposes)
	expectPrintedLocal(t, ".foo:hover { composes: bar; color: red }", ".foo:hover {\n  color: red;\n}\n", badComposes)
	expectPrintedLocal(t, ".foo[href] { composes: bar; color: red }", ".foo[href] {\n  color: red;\n}\n", badComposes)
	expectPrintedLocal(t, ".foo .bar { composes: bar; color: red }", ".foo .bar {\n  color: red;\n}\n", badComposes)
	expectPrintedLocal(t, ".foo, div { composes: bar; color: red }", ".foo,\ndiv {\n  color: red;\n}\n", badComposes)
	expectPrintedLocal(t, ".foo { .bar { composes: foo; color: red } }", ".foo {\n  .bar {\n    color: red;\n  }\n}\n", badComposes)
}

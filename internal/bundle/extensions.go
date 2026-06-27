package bundle

// BundledExtension is shipped under Extensions/{ID}/ and loaded via --load-extension.
type BundledExtension struct {
	ID   string
	Name string
}

// BundledExtensions are loaded at startup with --load-extension.
var BundledExtensions = []BundledExtension{
	{ID: "ekdpkppgmlalfkphpibadldikjimijon", Name: "copytables"},
	{ID: "kdnbphngjjojcnnapaegdgjpgadlhbke", Name: "constellation-mix"},
}

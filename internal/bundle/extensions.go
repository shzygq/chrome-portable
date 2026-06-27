package bundle

// BundledExtension is shipped under Extensions/{ID}/ and installed into the profile at build time.
type BundledExtension struct {
	ID   string
	Name string
}

// BundledExtensions are installed into the profile at build time via external CRX
// registration (Data/External Extensions/*.json). Chrome 137+ removed --load-extension.
var BundledExtensions = []BundledExtension{
	{ID: "ekdpkppgmlalfkphpibadldikjimijon", Name: "copytables"},
	{ID: "kdnbphngjjojcnnapaegdgjpgadlhbke", Name: "constellation-mix"},
}

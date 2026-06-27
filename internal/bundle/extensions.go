package bundle

// BundledExtension is shipped under Extensions/{ID}/ and installed into the profile at build time.
type BundledExtension struct {
	ID   string
	Name string
}

// BundledExtensions are installed via CDP Extensions.loadUnpacked during packaging.
// Chrome 137+ removed --load-extension from branded Stable builds.
var BundledExtensions = []BundledExtension{
	{ID: "ekdpkppgmlalfkphpibadldikjimijon", Name: "copytables"},
	{ID: "kdnbphngjjojcnnapaegdgjpgadlhbke", Name: "constellation-mix"},
}

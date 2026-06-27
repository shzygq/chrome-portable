package bundle

// BundledExtension is shipped unpacked under Extensions/{ID}/ for manual install.
type BundledExtension struct {
	ID   string
	Name string
}

// BundledExtensions are downloaded at build time into Extensions/{ID}/.
// Install them manually via chrome://extensions → Load unpacked.
var BundledExtensions = []BundledExtension{
	{ID: "ekdpkppgmlalfkphpibadldikjimijon", Name: "copytables"},
	{ID: "kdnbphngjjojcnnapaegdgjpgadlhbke", Name: "constellation-mix"},
}

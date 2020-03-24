package translation

var Translation = &PackageTranslation{
	Item:     &ItemTranslation{},
	KV:       &KVTranslation{},
	MassMail: &MassMailTranslation{},
}

type PackageTranslation struct {
	Item     *ItemTranslation
	KV       *KVTranslation
	MassMail *MassMailTranslation
}

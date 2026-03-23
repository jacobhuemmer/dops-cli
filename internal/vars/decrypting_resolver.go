package vars

import (
	"dops/internal/crypto"
	"dops/internal/domain"
)

type DecryptingVarResolver struct {
	inner VarResolver
	enc   crypto.Encrypter
}

func NewDecryptingResolver(inner VarResolver, enc crypto.Encrypter) *DecryptingVarResolver {
	return &DecryptingVarResolver{inner: inner, enc: enc}
}

func (r *DecryptingVarResolver) Resolve(cfg *domain.Config, catalogName, runbookName string, params []domain.Parameter) map[string]string {
	resolved := r.inner.Resolve(cfg, catalogName, runbookName, params)

	for k, v := range resolved {
		if crypto.IsEncrypted(v) {
			decrypted, err := r.enc.Decrypt(v)
			if err == nil {
				resolved[k] = decrypted
			}
		}
	}

	return resolved
}

var _ VarResolver = (*DecryptingVarResolver)(nil)

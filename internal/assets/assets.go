package assets

import (
	"time"

	"github.com/jessevdk/go-assets"
)


// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/cmodules/segmentation-kit/models": []string{"hmmdefs_monof_mix16_gid.binhmm"}, "/": []string{"cmodules"}, "/cmodules": []string{"segmentation-kit"}, "/cmodules/segmentation-kit": []string{"models"}}, map[string]*assets.File{
	"/cmodules/segmentation-kit": &assets.File{
		Path:     "/cmodules/segmentation-kit",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1532944384, 1532944384000000000),
		Data:     nil,
	}, "/cmodules/segmentation-kit/models": &assets.File{
		Path:     "/cmodules/segmentation-kit/models",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1532944384, 1532944384000000000),
		Data:     nil,
	}, "/cmodules/segmentation-kit/models/hmmdefs_monof_mix16_gid.binhmm": &assets.File{
		Path:     "/cmodules/segmentation-kit/models/hmmdefs_monof_mix16_gid.binhmm",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1532944384, 1532944384000000000),
		Data:     []byte(_Assets42d555a62090ed79d822d1c4a710ad09902ca9e8),
	}, "/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1532955343, 1532955343000000000),
		Data:     nil,
	}, "/cmodules": &assets.File{
		Path:     "/cmodules",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1532944326, 1532944326000000000),
		Data:     nil,
	}}, "")
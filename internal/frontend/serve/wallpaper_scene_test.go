package serve

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"sort"
	"testing"
)

func sceneTestPNG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0x20, G: 0x40, B: 0x80, A: 0xff})
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sceneTestJPEG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0x80, G: 0x40, B: 0x20, A: 0xff})
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// sceneTestPackage writes the PKG table exactly as Wallpaper Engine does:
// an 8-byte header, the magic, an entry count, then (name length, name,
// data offset, data length) for every file. Offsets are relative to the data
// section that follows the table.
func sceneTestPackage(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	var table, data bytes.Buffer
	table.Write([]byte{8, 0, 0, 0})
	table.WriteString("PKGV0024")
	if err := binary.Write(&table, binary.LittleEndian, uint32(len(names))); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if err := binary.Write(&table, binary.LittleEndian, uint32(len(name))); err != nil {
			t.Fatal(err)
		}
		table.WriteString(name)
		if err := binary.Write(&table, binary.LittleEndian, uint32(data.Len())); err != nil {
			t.Fatal(err)
		}
		if err := binary.Write(&table, binary.LittleEndian, uint32(len(files[name]))); err != nil {
			t.Fatal(err)
		}
		data.Write(files[name])
	}
	return append(table.Bytes(), data.Bytes()...)
}

func sceneTestProject() []byte {
	return []byte(`{"title":"Test Scene","type":"scene","file":"scene.json","preview":"preview.jpg"}`)
}

func sceneTestFiles(texture []byte) map[string][]byte {
	return map[string][]byte{
		"scene.json":          []byte(`{"general":{"cameraparallax":true,"cameraparallaxamount":0.03},"objects":[{"image":"models/main.json","sound":["sounds/song.mp3"]}]}`),
		"models/main.json":    []byte(`{"material":"materials/main.json"}`),
		"materials/main.json": []byte(`{"passes":[{"textures":["main"]}]}`),
		"materials/main.tex":  texture,
		"sounds/song.mp3":     []byte("ID3-test-audio"),
	}
}

func TestImportSceneWallpaperReadsTextureAudioAndParallax(t *testing.T) {
	wantImage := sceneTestPNG(t)
	pkg := sceneTestPackage(t, sceneTestFiles(append([]byte("TEXV0005\x00TEXI0001\x00"), wantImage...)))

	got, err := importSceneWallpaper(sceneTestProject(), pkg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Test Scene" {
		t.Errorf("title = %q, want Test Scene", got.Title)
	}
	if got.ImageMIME != "image/png" {
		t.Errorf("image mime = %q, want image/png", got.ImageMIME)
	}
	if !bytes.Equal(got.Image, wantImage) {
		t.Errorf("image bytes differ from the texture's PNG")
	}
	if got.AudioMIME != "audio/mpeg" || string(got.Audio) != "ID3-test-audio" {
		t.Errorf("audio = (%q, %q), want the mp3 beside the scene", got.AudioMIME, got.Audio)
	}
	if got.Parallax != 0.03 {
		t.Errorf("parallax = %v, want 0.03", got.Parallax)
	}
}

func TestImportSceneWallpaperFallsBackToPreview(t *testing.T) {
	preview := sceneTestJPEG(t)
	files := sceneTestFiles([]byte("TEXV0005\x00TEXI0001\x00 no image here"))
	pkg := sceneTestPackage(t, files)

	got, err := importSceneWallpaper(sceneTestProject(), pkg, preview)
	if err != nil {
		t.Fatal(err)
	}
	if got.ImageMIME != "image/jpeg" {
		t.Errorf("image mime = %q, want image/jpeg", got.ImageMIME)
	}
	if !bytes.Equal(got.Image, preview) {
		t.Errorf("image bytes differ from the preview")
	}
}

func TestImportSceneWallpaperNeedsAnImage(t *testing.T) {
	pkg := sceneTestPackage(t, sceneTestFiles([]byte("TEXV0005\x00TEXI0001\x00 no image here")))
	if _, err := importSceneWallpaper(sceneTestProject(), pkg, nil); err == nil {
		t.Fatal("a scene with no texture and no preview was accepted")
	}
}

func TestParseWallpaperPackageRejectsATruncatedEntry(t *testing.T) {
	pkg := sceneTestPackage(t, sceneTestFiles([]byte("TEXV0005\x00TEXI0001\x00")))
	if _, err := parseWallpaperPackage(pkg[:len(pkg)-1]); err == nil {
		t.Fatal("a truncated package was accepted")
	}
}

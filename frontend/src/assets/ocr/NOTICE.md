# Bundled local document processors

Chartrove serves these dependency assets locally. Versions and integrity hashes
are recorded in `frontend/yarn.lock`; binaries and language data are copied from
installed packages by Angular, not committed or fetched by the browser from a CDN.

- Tesseract.js 7.0.0: https://github.com/naptha/tesseract.js — Apache-2.0.
- Tesseract.js-core 7.0.0: https://github.com/naptha/tesseract.js-core — Apache-2.0;
  includes Tesseract/Leptonica/WebAssembly components and their upstream notices.
- PDF.js (`pdfjs-dist`) 6.5.438: https://github.com/mozilla/pdf.js — Apache-2.0;
  font and decoder directories retain the upstream licenses distributed with them.
- `@tesseract.js-data/eng` 1.0.0: https://github.com/naptha/tessdata — packaged English
  trained data, `4.0.0_best_int/eng.traineddata.gz`. The package declares MIT; the
  underlying Tesseract trained-data project is https://github.com/tesseract-ocr/tessdata
  (Apache-2.0). Preserve those upstream attributions when redistributing.

Unmodified package license files are published in `assets/ocr/licenses/`; PDF.js
font/decoder license files accompany their asset directories. Chartrove's worker
bootstrap is a lifecycle wrapper; recognition and PDF parsing use public library APIs.

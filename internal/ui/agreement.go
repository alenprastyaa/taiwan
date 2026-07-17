package ui

import (
	"strings"

	"university_agency/internal/dashboard"
)

// studentAgreementPage gates the rest of the student dashboard: unsigned
// students see the full agreement text + checkbox + typed-name form; signed
// students see a confirmation and a PDF download link as proof.
func studentAgreementPage(vm dashboard.ViewModel) string {
	if vm.HasSignedAgreement {
		agreement := vm.Agreement
		return `<div class="panel">
  <div class="auth-notice">Kamu sudah menyetujui perjanjian ini pada ` + dateLabel(agreement.AgreedAt) + `.</div>
  <h2 class="text-lg font-semibold mt-3">Surat Perjanjian</h2>
  <p class="text-sm text-slate-500 mt-1">Ditandatangani oleh: <strong>` + esc(agreement.FullNameTyped) + `</strong></p>
  <a class="primary-button mt-4" href="/student/agreement/pdf">Download PDF Perjanjian Saya</a>
</div>`
	}
	return `<div class="panel">
  <h2 class="text-lg font-semibold">Surat Perjanjian Kerjasama Layanan</h2>
  <p class="text-sm text-slate-500 mb-3">Baca dan setujui perjanjian ini terlebih dahulu untuk mengakses fitur lain di dashboard kamu.</p>
  <div class="agreement-text">` + nl2p(dashboard.AgreementText) + `</div>
  <form method="post" action="/student/agreement/sign" class="student-upload-form mt-4">
    <label class="checkbox-label"><input type="checkbox" name="agree" value="1" required> Saya telah membaca dan menyetujui perjanjian ini</label>
    <label>Nama Lengkap (sebagai tanda tangan)<input name="full_name" required placeholder="Ketik nama lengkap kamu"></label>
    <button class="primary-button" type="submit">Setujui &amp; Lanjutkan</button>
  </form>
</div>`
}

func nl2p(text string) string {
	var out strings.Builder
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			out.WriteString("<br>")
			continue
		}
		out.WriteString("<p>" + esc(line) + "</p>")
	}
	return out.String()
}

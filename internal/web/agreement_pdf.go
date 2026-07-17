package web

import (
	"bytes"
	"strings"

	"university_agency/internal/dashboard"

	"github.com/go-pdf/fpdf"
)

// renderAgreementPDF builds a one-page signed-agreement record: letterhead,
// the agreement text, and a signature block (typed name, timestamp, IP,
// user agent) as legal proof for both parties. Mirrors renderInvoicePDF's
// structural style (same letterhead/color pattern) for visual consistency.
func renderAgreementPDF(appName string, agreement dashboard.ClientAgreement) []byte {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()
	pdf.SetTitle("Surat Perjanjian "+agreement.ClientName, false)

	navy := func() { pdf.SetTextColor(7, 18, 37) }
	gray := func() { pdf.SetTextColor(100, 116, 139) }
	dark := func() { pdf.SetTextColor(15, 23, 42) }

	pdf.SetXY(20, 20)
	pdf.SetFont("Helvetica", "B", 18)
	navy()
	pdf.CellFormat(120, 8, strings.ToUpper(appName), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 16)
	gray()
	pdf.CellFormat(50, 8, "PERJANJIAN", "", 0, "R", false, 0, "")

	pdf.SetXY(20, 29)
	pdf.SetFont("Helvetica", "", 10)
	gray()
	pdf.CellFormat(170, 6, "Bukti Persetujuan Perjanjian Kerjasama Layanan", "", 0, "L", false, 0, "")

	pdf.SetDrawColor(226, 232, 240)
	pdf.SetLineWidth(0.4)
	pdf.Line(20, 40, 190, 40)

	pdf.SetXY(20, 48)
	pdf.SetFont("Helvetica", "B", 12)
	dark()
	pdf.CellFormat(170, 7, agreement.ClientName, "", 2, "L", false, 0, "")

	pdf.SetX(20)
	pdf.SetFont("Helvetica", "", 10)
	dark()
	pdf.MultiCell(170, 5.2, agreement.AgreementText, "", "L", false)

	pdf.SetY(pdf.GetY() + 8)
	pdf.SetDrawColor(226, 232, 240)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())

	signY := pdf.GetY() + 6
	pdf.SetXY(20, signY)
	pdf.SetFont("Helvetica", "B", 9)
	gray()
	pdf.CellFormat(0, 5, "BUKTI TANDA TANGAN ELEKTRONIK", "", 2, "L", false, 0, "")

	signRow := func(label, value string) {
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "", 9)
		gray()
		pdf.CellFormat(45, 6, label, "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		dark()
		pdf.CellFormat(125, 6, value, "", 2, "L", false, 0, "")
	}
	signRow("Ditandatangani oleh", agreement.FullNameTyped)
	signRow("Pada tanggal", agreement.AgreedAt.Format("02 Jan 2006 15:04")+" WIB")
	signRow("Versi Perjanjian", agreement.AgreementVersion)
	signRow("Alamat IP", agreement.IPAddress)
	signRow("User Agent", agreement.UserAgent)

	var buf bytes.Buffer
	_ = pdf.Output(&buf)
	return buf.Bytes()
}

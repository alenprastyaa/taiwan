package web

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"university_agency/internal/dashboard"

	"github.com/go-pdf/fpdf"
)

// renderInvoicePDF builds a standard one-page office invoice: letterhead,
// bill-to and invoice meta side by side, a line-item table, a totals block,
// and payment instructions in the footer. Replaces the earlier hand-rolled
// PDF that was just eight lines of plain left-aligned text.
//
// payments is the full payment log (any order) — filtered down to this
// order's verified entries to derive a "Metode Pembayaran" label (e.g.
// "Cicilan 3x") from real history, since there's no separate installment
// *plan* concept to read a schedule from.
func renderInvoicePDF(appName string, order dashboard.Order, payments []dashboard.OrderPayment) []byte {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()
	// Fixed single-page layout: disable auto page-break so a footer pinned near
	// the bottom (SetY with a negative value) never spills onto a second page.
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetTitle("Invoice "+order.Code, false)

	remaining := order.Total - order.Paid
	if remaining < 0 {
		remaining = 0
	}

	verifiedCount := 0
	for _, payment := range payments {
		if payment.OrderID == order.ID && payment.Status == dashboard.PaymentVerified {
			verifiedCount++
		}
	}
	methodLabel := "Belum Bayar"
	switch {
	case verifiedCount == 1:
		methodLabel = "Sekali Bayar"
	case verifiedCount > 1:
		methodLabel = "Cicilan " + strconv.Itoa(verifiedCount) + "x"
	}
	payoffLabel := "Belum Lunas"
	if order.PaidAt != nil {
		payoffLabel = order.PaidAt.Format("02 Jan 2006")
	}

	navy := func() { pdf.SetTextColor(7, 18, 37) }
	gray := func() { pdf.SetTextColor(100, 116, 139) }
	dark := func() { pdf.SetTextColor(15, 23, 42) }

	// Letterhead: company name + tagline on the left, "INVOICE" + code on the right.
	pdf.SetXY(20, 20)
	pdf.SetFont("Helvetica", "B", 18)
	navy()
	pdf.CellFormat(100, 8, strings.ToUpper(appName), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 22)
	gray()
	pdf.CellFormat(70, 8, "INVOICE", "", 0, "R", false, 0, "")

	pdf.SetXY(20, 29)
	pdf.SetFont("Helvetica", "", 10)
	gray()
	pdf.CellFormat(100, 6, "Taiwan Education Consultant", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 11)
	navy()
	pdf.CellFormat(70, 6, "#"+order.Code, "", 0, "R", false, 0, "")

	pdf.SetDrawColor(226, 232, 240)
	pdf.SetLineWidth(0.4)
	pdf.Line(20, 40, 190, 40)

	// Bill-to block on the left, invoice metadata on the right.
	leftY := 48.0
	pdf.SetXY(20, leftY)
	pdf.SetFont("Helvetica", "B", 9)
	gray()
	pdf.CellFormat(90, 5, "DITAGIHKAN KEPADA", "", 0, "L", false, 0, "")

	pdf.SetXY(20, leftY+7)
	pdf.SetFont("Helvetica", "B", 13)
	dark()
	pdf.CellFormat(90, 7, order.ClientName, "", 0, "L", false, 0, "")

	pdf.SetXY(20, leftY+15)
	pdf.SetFont("Helvetica", "", 10)
	gray()
	pdf.CellFormat(90, 6, order.PackageName, "", 0, "L", false, 0, "")

	metaRow := func(y float64, label, value string, r, g, b int) {
		pdf.SetXY(115, y)
		pdf.SetFont("Helvetica", "", 9)
		gray()
		pdf.CellFormat(35, 6, label, "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(r, g, b)
		pdf.CellFormat(40, 6, value, "", 0, "R", false, 0, "")
	}
	statusLabel, statusR, statusG, statusB := invoiceStatusStyle(order.Status, order.Paid)
	metaRow(leftY, "Tanggal Terbit", order.CreatedAt.Format("02 Jan 2006"), 15, 23, 42)
	metaRow(leftY+7, "Metode Bayar", methodLabel, 15, 23, 42)
	metaRow(leftY+14, "Tgl Pelunasan", payoffLabel, 15, 23, 42)
	metaRow(leftY+21, "Status", statusLabel, statusR, statusG, statusB)

	// Line-item table: shaded header row, one bordered data row.
	tableY := leftY + 34
	pdf.SetXY(20, tableY)
	pdf.SetFillColor(248, 250, 252)
	gray()
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(120, 9, "  DESKRIPSI", "", 0, "L", true, 0, "")
	pdf.CellFormat(50, 9, "JUMLAH (IDR)  ", "", 0, "R", true, 0, "")

	pdf.SetXY(20, tableY+9)
	pdf.SetFont("Helvetica", "", 10)
	dark()
	pdf.CellFormat(120, 10, "  "+order.PackageName, "B", 0, "L", false, 0, "")
	pdf.CellFormat(50, 10, formatIDR(order.Total)+"  ", "B", 0, "R", false, 0, "")

	// Totals block, right-aligned under the table.
	totalsY := tableY + 24
	totalsX := 110.0
	rowH := 7.0
	row := 0
	totalRow := func(label, value string, bold bool) {
		y := totalsY + float64(row)*rowH
		pdf.SetXY(totalsX, y)
		if bold {
			pdf.SetFont("Helvetica", "B", 11)
			dark()
		} else {
			pdf.SetFont("Helvetica", "", 10)
			gray()
		}
		pdf.CellFormat(45, rowH, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(35, rowH, value, "", 0, "R", false, 0, "")
		row++
	}
	totalRow("Subtotal", formatIDR(order.Total), false)
	if order.Paid > 0 {
		totalRow("Sudah Dibayar", "-"+formatIDR(order.Paid), false)
	}
	lineY := totalsY + float64(row)*rowH + 2
	pdf.Line(totalsX, lineY, 190, lineY)
	row++
	balanceLabel := "Sisa Tagihan"
	if remaining == 0 {
		balanceLabel = "Total Lunas"
	}
	totalRow(balanceLabel, formatIDR(remaining), true)
	if remaining > 0 {
		nextY := totalsY + float64(row)*rowH
		pdf.SetXY(totalsX, nextY)
		pdf.SetFont("Helvetica", "", 10)
		gray()
		pdf.CellFormat(45, 4.5, "Tagihan Berikutnya", "", 0, "L", false, 0, "")
		pdf.SetXY(totalsX, nextY+4.5)
		pdf.SetFont("Helvetica", "", 7)
		pdf.CellFormat(45, 3, "jatuh tempo "+order.DueDate.Format("02 Jan 2006"), "", 0, "L", false, 0, "")
		pdf.SetXY(totalsX+45, nextY)
		pdf.SetFont("Helvetica", "B", 10)
		gray()
		pdf.CellFormat(35, 4.5, formatIDR(remaining), "", 0, "R", false, 0, "")
		row++
	}

	// Payment proof, if the client already submitted one.
	noteY := totalsY + float64(row)*rowH + 12
	if order.ProofNote != "" || order.ProofFileName != "" {
		pdf.SetXY(20, noteY)
		pdf.SetFont("Helvetica", "B", 9)
		gray()
		pdf.CellFormat(0, 6, "BUKTI PEMBAYARAN", "", 2, "L", false, 0, "")
		if order.ProofFileName != "" {
			pdf.SetX(20)
			pdf.SetFont("Helvetica", "", 9)
			dark()
			pdf.CellFormat(0, 5, "File: "+order.ProofFileName, "", 2, "L", false, 0, "")
		}
		if order.ProofNote != "" {
			pdf.SetX(20)
			pdf.SetFont("Helvetica", "", 9)
			dark()
			pdf.MultiCell(170, 5, "Catatan: "+order.ProofNote, "", "L", false)
		}
	}

	// Footer: payment instructions, pinned near the bottom of the page.
	pdf.SetY(-35)
	pdf.SetDrawColor(226, 232, 240)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.SetXY(20, pdf.GetY()+4)
	pdf.SetFont("Helvetica", "", 8)
	gray()
	pdf.MultiCell(170, 4.5, "Pembayaran dilakukan melalui transfer manual. Setelah transfer, unggah bukti pembayaran melalui portal client untuk diverifikasi.\nInvoice ini dibuat otomatis oleh sistem pada "+time.Now().Format("02 Jan 2006 15:04")+" WIB.", "", "L", false)

	var buf bytes.Buffer
	_ = pdf.Output(&buf)
	return buf.Bytes()
}

func invoiceStatusStyle(status dashboard.OrderStatus, paid int64) (label string, r, g, b int) {
	switch {
	case status == dashboard.OrderPaid:
		return "LUNAS", 5, 150, 105
	case status == dashboard.OrderWaitingVerification:
		return "MENUNGGU VERIFIKASI", 180, 83, 9
	case paid > 0:
		return "CICILAN MASUK", 29, 78, 216
	default:
		return "BELUM DIBAYAR", 185, 28, 28
	}
}

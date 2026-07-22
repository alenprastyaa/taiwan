package ui

import (
	"strings"

	"university_agency/internal/dashboard"
)

func shipmentDirectionLabel(direction dashboard.ShipmentDirection) string {
	if direction == dashboard.ShipmentIncoming {
		return "Client/Institusi → Agensi"
	}
	return "Agensi → Client/Institusi"
}

func shipmentStatusLabel(status dashboard.ShipmentStatus) string {
	if status == dashboard.ShipmentDelivered {
		return "Diterima"
	}
	return "Dikirim"
}

func shipmentStatusClass(status dashboard.ShipmentStatus) string {
	if status == dashboard.ShipmentDelivered {
		return "green"
	}
	return "amber"
}

func courierOptions(couriers []string) string {
	var opts strings.Builder
	for _, courier := range couriers {
		opts.WriteString(`<option value="` + attr(courier) + `">`)
	}
	return opts.String()
}

// staffLogistics is the owner/staff "Logistik Dokumen" page: a toggle-create
// form (mirrors staffExpenses) plus a table of every shipment visible to the
// viewer, with a "Tandai Diterima" action per row.
func staffLogistics(vm dashboard.ViewModel) string {
	basePath := "/" + vm.Role.String() + "/logistics"
	form := ""
	if vm.ShowCreateForm {
		form = `<form method="post" action="/logistics/create" class="student-upload-form">
  <label>Client<select name="client_id" required><option value="">Pilih client</option>` + clientSelectOptions(vm.Clients) + `</select></label>
  <label>Arah Pengiriman<select name="direction" required>
    <option value="` + string(dashboard.ShipmentOutgoing) + `">Agensi → Client/Institusi</option>
    <option value="` + string(dashboard.ShipmentIncoming) + `">Client/Institusi → Agensi</option>
  </select></label>
  <label>Ekspedisi<input name="courier" list="shipment-courier-list" required placeholder="Pilih atau ketik ekspedisi baru" autocomplete="off"><datalist id="shipment-courier-list">` + courierOptions(vm.ShipmentCouriers) + `</datalist></label>
  <label>Nomor Resi<input name="tracking_number" placeholder="JP1234567890ID"></label>
  <label class="span-2">Isi Dokumen<input name="contents" placeholder="Contoh: Ijazah asli, Transkrip Nilai"></label>
  <label>Alamat Pengirim<input name="sender_address" placeholder="Alamat asal pengiriman"></label>
  <label>Alamat Penerima<input name="recipient_address" placeholder="Alamat tujuan pengiriman"></label>
  <label>Tanggal Kirim<input name="shipped_date_label" placeholder="25/06/2026"></label>
  <button class="primary-button" type="submit">Simpan Pengiriman</button>
</form>`
	}
	return `<div class="panel"><div class="toolbar-row mb-4"><div><h2 class="text-lg font-semibold">Logistik Dokumen</h2><p class="text-sm text-slate-500">Catat dan lacak pengiriman dokumen ke/dari client, lengkap dengan resi dan ekspedisi.</p></div>` + toggleFormButton(vm.ShowCreateForm, "+ Catat Pengiriman", basePath) + `</div>` + form + shipmentsTable(vm.Shipments, true) + `</div>`
}

// studentLogistics is the client's read-only view of their own shipments.
func studentLogistics(vm dashboard.ViewModel) string {
	return `<div class="panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Logistik Dokumen</h2><p class="text-sm text-slate-500">Status pengiriman dokumen kamu, termasuk nomor resi dan ekspedisi.</p></div></div>
  ` + shipmentsTable(vm.Shipments, false) + `
</div>`
}

func shipmentsTable(shipments []dashboard.Shipment, actionable bool) string {
	var rows strings.Builder
	for _, shipment := range shipments {
		action := ""
		if actionable {
			action = `<button class="icon-action small" type="button" disabled>Sudah Diterima</button>`
			if shipment.Status != dashboard.ShipmentDelivered {
				action = `<form method="post" action="/logistics/` + attr(shipment.ID) + `/received" class="inline-form"><button class="icon-action" type="submit">Tandai Diterima</button></form>`
			}
		}
		address := esc(shipment.SenderAddress) + ` → ` + esc(shipment.RecipientAddress)
		received := shipment.ReceivedDateLabel
		if received == "" {
			received = "-"
		}
		row := `<tr><td><strong>` + esc(shipment.ClientName) + `</strong></td><td>` + esc(shipmentDirectionLabel(shipment.Direction)) + `</td><td>` + esc(shipment.Courier) + `</td><td>` + esc(shipment.TrackingNumber) + `</td><td>` + esc(shipment.Contents) + `</td><td>` + address + `</td><td><span class="status ` + shipmentStatusClass(shipment.Status) + `">` + shipmentStatusLabel(shipment.Status) + `</span></td><td>` + esc(shipment.ShippedDateLabel) + `</td><td>` + esc(received) + `</td>`
		if actionable {
			row += `<td>` + action + `</td>`
		}
		row += `</tr>`
		rows.WriteString(row)
	}
	columns := 9
	if actionable {
		columns = 10
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="` + intText(columns) + `">Belum ada pengiriman tercatat.</td></tr>`)
	}
	actionHeader := ""
	if actionable {
		actionHeader = `<th>Aksi</th>`
	}
	return `<table class="data-table"><thead><tr><th>Client</th><th>Arah</th><th>Ekspedisi</th><th>No. Resi</th><th>Isi Dokumen</th><th>Alamat</th><th>Status</th><th>Tgl Kirim</th><th>Tgl Terima</th>` + actionHeader + `</tr></thead><tbody>` + rows.String() + `</tbody></table>`
}

import frappe
from frappe.model.document import Document

class Asset(Document):
	def autoname(self):
		# Fiat: asset_code = currency
		if self.type == "Fiat" and self.currency:
			self.asset_code = self.currency
		# Commodity: asset_code = commodity_name
		elif self.type == "Commodity" and self.commodity_name:
			self.asset_code = self.commodity_name

	def validate(self):
		if self.type == "Fiat":
			if not self.currency:
				frappe.throw("Для типа Fiat выберите Currency")
			self.asset_code = self.currency

		elif self.type == "Crypto":
			self.currency = None
			if not self.asset_code:
				frappe.throw("Для типа Crypto укажите Asset Code")

		elif self.type == "Commodity":
			if not self.commodity_name:
				frappe.throw("Для типа Commodity заполните Commodity Name")
			self.asset_code = self.commodity_name
			self.currency = None
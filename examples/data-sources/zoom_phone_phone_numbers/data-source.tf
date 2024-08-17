data "zoom_phone_phone_numbers" "all" {}

data "zoom_phone_phone_numbers" "filtered" {
  filter = {
    type = "unassigned"
  }
}

output "example" {
  value = data.zoom_phone_phone_numbers.all
}

output "filtered" {
  value = data.zoom_phone_phone_numbers.filtered
}

output "search_phone_number" {
  value = try(data.zoom_phone_phone_numbers.all.phone_numbers[index(data.zoom_phone_phone_numbers.all.phone_numbers.*.number, "+1234567890")], null)
}

document.addEventListener("click", async (event) => {
  const button = event.target.closest("[data-copy]");
  if (!button) return;
  const input = document.getElementById(button.dataset.copy);
  const status = document.getElementById("copy-status");
  try {
    await navigator.clipboard.writeText(input.value);
    status.textContent = "Đã sao chép lời nhắc.";
  } catch {
    input.focus();
    input.select();
    status.textContent = "Đã chọn lời nhắc. Hãy dùng thao tác sao chép trên thiết bị.";
  }
});
document.querySelector("[data-room-search]")?.addEventListener("input", (event) => {
  const query = event.target.value.trim().toLocaleLowerCase("vi");
  for (const room of document.querySelectorAll("[data-room]")) {
    room.hidden = !room.dataset.room.toLocaleLowerCase("vi").includes(query);
  }
});

document.addEventListener("DOMContentLoaded", function () {
    var sidebarToggle = document.getElementById("sidebarToggle");
    if (sidebarToggle) {
        sidebarToggle.addEventListener("click", function (e) {
            e.preventDefault();
            document.getElementById("wrapper").classList.toggle("toggled");
        });
    }
});

function resetUserForm() {
    document.getElementById("formUser").action = "/admin/users";
    document.getElementById("modalUserTitle").innerHTML = '<i class="bi bi-person-plus"></i> Tambah User';
    document.getElementById("userID").value = "";
    document.getElementById("userName").value = "";
    document.getElementById("userUsername").value = "";
    document.getElementById("userPassword").value = "";
    document.getElementById("userPassword").required = true;
    document.getElementById("passwordHint").textContent = "(wajib diisi)";
    document.getElementById("userRole").value = "user";
}

function editUser(id, name, username, role) {
    document.getElementById("formUser").action = "/admin/users/update/" + id;
    document.getElementById("modalUserTitle").innerHTML = '<i class="bi bi-pencil"></i> Edit User';
    document.getElementById("userID").value = id;
    document.getElementById("userName").value = name;
    document.getElementById("userUsername").value = username;
    document.getElementById("userPassword").value = "";
    document.getElementById("userPassword").required = false;
    document.getElementById("passwordHint").textContent = "(kosongkan jika tidak diubah)";
    document.getElementById("userRole").value = role;
}

function editEbupot(id, userId, bulan, tahun) {
    document.getElementById("formEdit").action = "/admin/ebupots/update/" + id;
    document.getElementById("editUserID").value = userId;
    document.getElementById("editBulan").value = bulan;
    document.getElementById("editTahun").value = tahun;
}

var contracts = []
var currentContractId = null

function newContract() {
    currentContractId = null
    $("#contractModalLabel").text("New Contract")
    $("#contractId").val("")
    $("#contractName").val("")
    $("#contractClient").val("")
    $("#contractStatus").val("draft")
    $("#approverTable tbody").empty()
    $("#versionTable tbody").empty()
}

function edit(id) {
    if (id === -1) {
        newContract()
        return
    }
    currentContractId = id
    var c = contracts.find(function (c) { return c.id === id })
    $("#contractModalLabel").text("Edit Contract: " + c.name)
    $("#contractId").val(c.id)
    $("#contractName").val(c.name)
    $("#contractClient").val(c.client_name)
    $("#contractStatus").val(c.status)
    renderApprovers(c.approvers || [])
    renderVersions(c.versions || [])
}

function renderApprovers(approvers) {
    var tbody = $("#approverTable tbody")
    tbody.empty()
    approvers.forEach(function (a) {
        tbody.append("<tr><td>" + escapeHtml(a.name) + "</td><td>" + escapeHtml(a.email) +
            "</td><td><button type=\"button\" class=\"btn btn-danger btn-xs\" onclick=\"removeApprover(" + a.id + ")\"><i class=\"fa fa-trash-o\"></i></button></td></tr>")
    })
}

function renderVersions(versions) {
    var tbody = $("#versionTable tbody")
    tbody.empty()
    versions.forEach(function (v) {
        tbody.append("<tr><td>" + v.version_number + "</td><td>" + escapeHtml(v.file_name) +
            "</td><td>" + escapeHtml(v.scope_description || "") + "</td><td>" + v.uploaded_at + "</td>" +
            "<td><button type=\"button\" class=\"btn btn-primary btn-xs\" onclick=\"issueApprovalForVersion(" + v.id + ")\">Request Approval</button></td></tr>")
    })
}

function addApprover() {
    var name = $("#newApproverName").val()
    var email = $("#newApproverEmail").val()
    if (!name || !email) {
        errorFlash("Approver name and email are required")
        return
    }
    if (!currentContractId) {
        errorFlash("Save the contract before adding approvers")
        return
    }
    api.contractId.addApprover(currentContractId, { name: name, email: email }).success(function () {
        $("#newApproverName").val("")
        $("#newApproverEmail").val("")
        edit(currentContractId)
        load()
    }).error(function () {
        errorFlash("Error adding approver")
    })
}

function removeApprover(approverId) {
    api.contractId.deleteApprover(currentContractId, approverId).success(function () {
        edit(currentContractId)
        load()
    })
}

function uploadVersion() {
    if (!currentContractId) {
        errorFlash("Save the contract before uploading a document")
        return
    }
    var fileInput = document.getElementById("versionFile")
    if (!fileInput.files.length) {
        errorFlash("Choose a file to upload")
        return
    }
    var formData = new FormData()
    formData.append("file", fileInput.files[0])
    formData.append("scope_description", $("#versionScope").val())
    $.ajax({
        url: "/api/contracts/" + currentContractId + "/versions",
        method: "POST",
        data: formData,
        processData: false,
        contentType: false,
        beforeSend: function (xhr) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + user.api_key)
        }
    }).done(function () {
        successFlash("Contract version uploaded")
        fileInput.value = ""
        $("#versionScope").val("")
        edit(currentContractId)
        load()
    }).fail(function () {
        errorFlash("Error uploading contract version")
    })
}

function issueApprovalForVersion(versionId) {
    api.approvals.issue({ contract_version_id: versionId }).success(function (data) {
        if (data.approvers_notified < data.approvers_total) {
            errorFlash("Approval request created, but only " + data.approvers_notified + " of " +
                data.approvers_total + " approver(s) could be e-mailed. Check the sending profile configuration.")
        } else {
            successFlash("Approval request sent to configured approvers")
        }
    }).error(function (data) {
        errorFlash(data.responseJSON ? data.responseJSON.message : "Error issuing approval request")
    })
}

function save() {
    var contract = {
        name: $("#contractName").val(),
        client_name: $("#contractClient").val(),
        status: $("#contractStatus").val()
    }
    var id = $("#contractId").val()
    var request
    if (id) {
        contract.id = parseInt(id)
        request = api.contractId.put(contract)
    } else {
        request = api.contracts.post(contract)
    }
    request.success(function (data) {
        successFlash("Contract saved successfully!")
        currentContractId = data.id
        $("#contractId").val(data.id)
        load()
    }).error(function () {
        errorFlash("Error saving contract")
    })
}

function deleteContract(id) {
    if (!confirm("Delete this contract? This also removes its versions and approvers.")) {
        return
    }
    api.contractId.delete(id).success(function () {
        successFlash("Contract deleted")
        load()
    })
}

function load() {
    $("#loading").show()
    $("#contractTable").hide()
    api.contracts.get().success(function (cs) {
        $("#loading").hide()
        contracts = cs
        if (cs.length === 0) {
            $("#emptyMessage").show()
            return
        }
        $("#emptyMessage").hide()
        $("#contractTable").show()
        if ($.fn.DataTable.isDataTable("#contractTable")) {
            $("#contractTable").DataTable().destroy()
        }
        var rows = cs.map(function (c) {
            return [
                escapeHtml(c.name),
                escapeHtml(c.client_name),
                c.status,
                (c.versions || []).length,
                (c.approvers || []).length,
                "<div class=\"pull-right\">" +
                "<button type=\"button\" class=\"btn btn-primary\" data-toggle=\"modal\" data-target=\"#contractModal\" onclick=\"edit(" + c.id + ")\"><i class=\"fa fa-pencil\"></i></button>&nbsp;" +
                "<button type=\"button\" class=\"btn btn-danger\" onclick=\"deleteContract(" + c.id + ")\"><i class=\"fa fa-trash-o\"></i></button>" +
                "</div>"
            ]
        })
        $("#contractTable").DataTable({
            data: rows,
            destroy: true,
            columnDefs: [{ orderable: false, targets: "no-sort" }]
        })
    }).error(function () {
        $("#loading").hide()
        errorFlash("Error fetching contracts")
    })
}

$(document).ready(function () {
    load()
})

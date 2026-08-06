var currentApprovalId = null

var statusLabels = {
    pending: "label-warning",
    approved: "label-success",
    rejected: "label-danger",
    changes_requested: "label-info",
    expired: "label-default"
}

function statusBadge(status) {
    var cls = statusLabels[status] || "label-default"
    return "<span class=\"label " + cls + "\">" + status + "</span>"
}

function resend(id) {
    api.approvalId.resend(id).success(function (data) {
        if (data.approvers_notified < data.approvers_total) {
            errorFlash("Reminder recorded, but only " + data.approvers_notified + " of " +
                data.approvers_total + " approver(s) could be e-mailed. Check the sending profile configuration.")
        } else {
            successFlash("Approval reminder resent")
        }
        load()
    }).error(function () {
        errorFlash("Error resending approval request")
    })
}

function viewApproval(id) {
    currentApprovalId = id
    api.approvalId.get(id).success(function (ar) {
        $("#approvalModalLabel").text("Approval Request #" + ar.id)
        $("#approvalStatus").html("Status: " + statusBadge(ar.status) +
            "<br/>Requested: " + ar.requested_at +
            (ar.decided_at ? "<br/>Decided: " + ar.decided_at : ""))
        renderComments(ar.comments || [])
        $("#approvalModal").modal("show")
    })
}

function renderComments(comments) {
    var el = $("#approvalComments")
    el.empty()
    if (comments.length === 0) {
        el.append("<p class=\"text-muted\">No comments yet.</p>")
        return
    }
    comments.forEach(function (c) {
        el.append("<div style=\"margin-bottom:8px;\"><strong>" + escapeHtml(c.author_type) + "</strong> - " +
            escapeHtml(c.created_at) + "<br/>" + escapeHtml(c.body) + "</div>")
    })
}

function addComment() {
    var body = $("#newComment").val()
    if (!body) {
        return
    }
    api.approvalId.comment(currentApprovalId, body).success(function () {
        $("#newComment").val("")
        viewApproval(currentApprovalId)
    })
}

function exportEvidence() {
    $.ajax({
        url: "/api/approvals/" + currentApprovalId + "/export",
        method: "GET",
        beforeSend: function (xhr) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + user.api_key)
        }
    }).done(function (data) {
        var blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" })
        var url = URL.createObjectURL(blob)
        var a = document.createElement("a")
        a.href = url
        a.download = "approval-" + currentApprovalId + "-evidence.json"
        a.click()
        URL.revokeObjectURL(url)
    }).fail(function () {
        errorFlash("Error exporting evidence")
    })
}

function load() {
    $("#loading").show()
    $("#approvalTable").hide()
    api.approvals.get().success(function (rows) {
        $("#loading").hide()
        if (rows.length === 0) {
            $("#emptyMessage").show()
            return
        }
        $("#emptyMessage").hide()
        $("#approvalTable").show()
        var data = rows.map(function (r) {
            return [
                escapeHtml(r.contract_name),
                escapeHtml(r.client_name),
                r.version_number,
                statusBadge(r.status),
                r.requested_at,
                r.token_expires_at,
                r.last_reminder_sent_at || "-",
                "<div class=\"pull-right\">" +
                "<button type=\"button\" class=\"btn btn-primary\" onclick=\"viewApproval(" + r.id + ")\"><i class=\"fa fa-eye\"></i></button>&nbsp;" +
                (r.status === "pending" ? "<button type=\"button\" class=\"btn btn-default\" onclick=\"resend(" + r.id + ")\"><i class=\"fa fa-refresh\"></i> Resend</button>" : "") +
                "</div>"
            ]
        })
        if ($.fn.DataTable.isDataTable("#approvalTable")) {
            $("#approvalTable").DataTable().destroy()
        }
        $("#approvalTable").DataTable({
            data: data,
            destroy: true,
            columnDefs: [{ orderable: false, targets: "no-sort" }]
        })
    }).error(function () {
        $("#loading").hide()
        errorFlash("Error fetching approval requests")
    })
}

$(document).ready(function () {
    load()
})

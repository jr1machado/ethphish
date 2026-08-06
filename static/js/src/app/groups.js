var groups = []

// Save attempts to POST or PUT to /groups/
function save(id) {
    var targets = []
    
    // Get targets from the table
    $.each($("#targetsTable").DataTable().rows().data(), function (i, target) {

        const email = unescapeHtml(target[2]).trim();
        const phone = unescapeHtml(target[3]).trim();
    
        // Skip if both are empty
        if (!email && !phone) return;

        targets.push({
            first_name: unescapeHtml(target[0]),
            last_name: unescapeHtml(target[1]),
            email: unescapeHtml(target[2]).trim().toLowerCase(),
            phone: unescapeHtml(target[3]).replace(/\D/g, ''),
            position: unescapeHtml(target[4]),
            custom: unescapeHtml(target[5]),
            department: unescapeHtml(target[6]),
            company: unescapeHtml(target[7]),
            city: unescapeHtml(target[8]),
            state: unescapeHtml(target[9]),
            country: unescapeHtml(target[10]),
            unit: unescapeHtml(target[11]),
            tags: unescapeHtml(target[12])
        })
    })
    
    // Get original targets from the hidden field if it exists
    // var originalTargetsJson = $("#original_targets").val();
    // if (originalTargetsJson && id != -1) {
    //     try {
    //         var originalTargets = JSON.parse(originalTargetsJson);
    //         console.log("Original targets:", originalTargets);
            
    //         // Merge targets, avoiding duplicates
    //         var mergedTargets = [...targets];
    //         var emailMap = {};
    //         var phoneMap = {};
            
    //         // Create maps of existing targets by email and phone
    //         targets.forEach(function(target) {
    //             if (target.email) {
    //                 const normalizedEmail = target.email.trim().toLowerCase();
    //                 emailMap[normalizedEmail] = true;
    //             }
    //             if (target.phone) {
    //                 const normalizedPhone = target.phone.replace(/\D/g, '');
    //                 phoneMap[normalizedPhone] = true;
    //             }
    //         });
            
    //         // Add original targets that aren't already in the table
    //         originalTargets.forEach(function(target) {
    //             var isDuplicate = false;
                
    //             // Check if this target is already in our list by email or phone
    //             if (target.email) {
    //                 const normalizedEmail = target.email.trim().toLowerCase();
    //                 if (emailMap[normalizedEmail]) {
    //                     isDuplicate = true;
    //                 }
    //             }
    //             if (target.phone) {
    //                 const normalizedPhone = target.phone.replace(/\D/g, '');
    //                 if (phoneMap[normalizedPhone]) {
    //                     isDuplicate = true;
    //                 }
    //             }
                
    //             // If not a duplicate, add it to the merged list
    //             if (!isDuplicate) {
    //                 mergedTargets.push(target);
    //             }
    //         });
            
    //         targets = mergedTargets;
    //     } catch (e) {
    //         console.error("Error parsing original targets:", e);
    //     }
    // }
    
    // Log the targets being sent to the server for debugging
    // console.log("Saving group with targets:", targets);
    
    var group = {
        name: $("#name").val(),
        targets: targets
    }
    // Submit the group
    if (id != -1) {
        // If we're just editing an existing group,
        // we need to PUT /groups/:id
        group.id = id
        api.groupId.put(group)
            .success(function (data) {
                successFlash("Group updated successfully!")
                load()
                dismiss()
                $("#modal").modal('hide')
            })
            .error(function (data) {
                // console.error("Error updating group:", data);
                modalError(data.responseJSON ? data.responseJSON.message : "An error occurred while updating the group")
            })
    } else {
        // Else, if this is a new group, POST it
        // to /groups
        api.groups.post(group)
            .success(function (data) {
                successFlash("Group added successfully!")
                load()
                dismiss()
                $("#modal").modal('hide')
            })
            .error(function (data) {
                // console.error("Error saving group:", data);
                modalError(data.responseJSON ? data.responseJSON.message : "An error occurred while saving the group")
            })
    }
}

function dismiss() {
    $("#targetsTable").dataTable().DataTable().clear().draw()
    $("#name").val("")
    $("#modal\\.flashes").empty()
}

function edit(id) {
    targets = $("#targetsTable").dataTable({
        destroy: true, // Destroy any other instantiated table - http://datatables.net/manual/tech-notes/3#destroy
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    })
    $("#modalSubmit").unbind('click').click(function () {
        save(id)
    })
    if (id == -1) {
        $("#groupModalLabel").text("New Group");
        // Don't try to fetch a group when creating a new one
        // Clear the table just in case
        targets.DataTable().clear().draw();
    } else {
        $("#groupModalLabel").text("Edit Group");
        api.groupId.get(id)
            .success(function (group) {
                // console.log("Group data received:", group);
                $("#name").val(group.name)
                targetRows = []
                if (group.targets && group.targets.length > 0) {
                    // console.log("Group targets:", group.targets);
                    $.each(group.targets, function (i, record) {
                      targetRows.push([
                          escapeHtml(record.first_name),
                          escapeHtml(record.last_name),
                          escapeHtml(record.email),
                          escapeHtml(record.phone),
                          escapeHtml(record.position),
                          escapeHtml(record.custom),
                          escapeHtml(record.department || ""),
                          escapeHtml(record.company || ""),
                          escapeHtml(record.city || ""),
                          escapeHtml(record.state || ""),
                          escapeHtml(record.country || ""),
                          escapeHtml(record.unit || ""),
                          escapeHtml(record.tags || ""),
                          '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
                      ])
                    });
                } else {
                    // console.warn("No targets found in group or targets array is empty");
                }
                targets.DataTable().rows.add(targetRows).draw()
                refreshFilterOptions();
                validateTargetRows();

                // Store the original targets in a hidden field so we can preserve them when saving
                $("#original_targets").val(JSON.stringify(group.targets));
            })
            .error(function () {
                errorFlash("Error fetching group")
            })
    }
    // Handle file uploads
    $("#csvupload").fileupload({
        url: "/api/import/group",
        dataType: "json",
        beforeSend: function (xhr) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + user.api_key);
        },
        add: function (e, data) {
            $("#modal\\.flashes").empty()
            var file = data.originalFiles[0];
            var filename = file ? file['name'] : "";
            var isXlsx = /\.xlsx$/i.test(filename);
            if (isXlsx) {
                var reader = new FileReader();
                reader.onload = function (evt) {
                    var workbook = XLSX.read(evt.target.result, { type: 'array' });
                    var firstSheetName = workbook.SheetNames[0];
                    var csvString = XLSX.utils.sheet_to_csv(workbook.Sheets[firstSheetName]);
                    var csvBlob = new Blob([csvString], { type: 'text/csv' });
                    var csvFile = new File([csvBlob], filename.replace(/\.xlsx$/i, '.csv'), { type: 'text/csv' });
                    data.files = [csvFile];
                    data.submit();
                };
                reader.onerror = function () {
                    modalError("Error reading XLSX file");
                };
                reader.readAsArrayBuffer(file);
                return;
            }
            var acceptFileTypes = /(csv|txt)$/i;
            if (filename && !acceptFileTypes.test(filename.split(".").pop())) {
                modalError("Unsupported file extension (use .csv, .txt, or .xlsx)")
                return false;
            }
            data.submit();
        },
        done: function (e, data) {
            $.each(data.result, function (i, record) {
                addTarget(
                    record.first_name,
                    record.last_name,
                    record.email,
                    record.phone,
                    record.position,
                    record.custom,
                    record.department,
                    record.company,
                    record.city,
                    record.state,
                    record.country,
                    record.unit,
                    record.tags);
            });
            targets.DataTable().draw();
            refreshFilterOptions();
            validateTargetRows();
        }
    })
}

var downloadCSVTemplate = function () {
    var csvScope = [{
        'First_Name': 'Example',
        'Last_Name': 'User',
        'Email': 'foobar@example.com',
        'Phone': '',
        'Position': 'Systems Administrator',
        'Custom': 'Custom value',
        'Department': 'Engineering',
        'Company': 'Acme Corp',
        'City': 'London',
        'State': 'England',
        'Country': 'UK',
        'Unit': 'Platform',
        'Tags': 'vip,executive'
    }, {
        'First_Name': 'Example2',
        'Last_Name': 'User2',
        'Email': '',
        'Phone': '+1234567890',
        'Position': 'Human Resources',
        'Custom': 'Foo bar',
        'Department': '',
        'Company': '',
        'City': '',
        'State': '',
        'Country': '',
        'Unit': '',
        'Tags': ''
    }]
    var filename = 'group_template.csv'
    var csvString = Papa.unparse(csvScope, {})
    var csvData = new Blob([csvString], {
        type: 'text/csv;charset=utf-8;'
    });
    if (navigator.msSaveBlob) {
        navigator.msSaveBlob(csvData, filename);
    } else {
        var csvURL = window.URL.createObjectURL(csvData);
        var dlLink = document.createElement('a');
        dlLink.href = csvURL;
        dlLink.setAttribute('download', filename)
        document.body.appendChild(dlLink)
        dlLink.click();
        document.body.removeChild(dlLink)
    }
}

var downloadGroup = function(id) {
    // Get the group details
    api.groupId.get(id)
        .success(function(group) {
            // Create CSV content with underscores in headers for easy re-upload
            var csvContent = "First_Name,Last_Name,Email,Phone,Position,Custom,Department,Company,City,State,Country,Unit,Tags\n";

            // Add each target to the CSV
            $.each(group.targets, function(i, target) {
                // Properly escape fields that might contain commas
                var firstName = escapeCsvField(target.first_name);
                var lastName = escapeCsvField(target.last_name);
                var email = escapeCsvField(target.email);
                var phone = escapeCsvField(target.phone);
                var position = escapeCsvField(target.position);
                var custom = escapeCsvField(target.custom);
                var department = escapeCsvField(target.department);
                var company = escapeCsvField(target.company);
                var city = escapeCsvField(target.city);
                var state = escapeCsvField(target.state);
                var country = escapeCsvField(target.country);
                var unit = escapeCsvField(target.unit);
                var tags = escapeCsvField(target.tags);

                // Add the row to CSV content
                csvContent += firstName + "," + lastName + "," + email + "," + phone + "," + position + "," + custom +
                    "," + department + "," + company + "," + city + "," + state + "," + country + "," + unit + "," + tags + "\n";
            });
            
            // Create a blob with the CSV content
            var blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            
            // Create a safe filename based on the group name
            var filename = "group_" + group.name.replace(/[^\w\-]+/g, '_').toLowerCase() + ".csv";
            
            // Handle different browser download methods
            if (navigator.msSaveBlob) { // IE 10+
                navigator.msSaveBlob(blob, filename);
            } else {
                // For other browsers
                var link = document.createElement("a");
                if (link.download !== undefined) { // Feature detection
                    // Create a URL for the blob
                    var url = URL.createObjectURL(blob);
                    link.setAttribute("href", url);
                    link.setAttribute("download", filename);
                    link.style.visibility = 'hidden';
                    document.body.appendChild(link);
                    link.click();
                    document.body.removeChild(link);
                    URL.revokeObjectURL(url);
                }
            }
            
            successFlash("Group exported successfully!");
        })
        .error(function() {
            errorFlash("Error fetching group data for export");
        });
}

// Helper function to escape CSV fields
function escapeCsvField(field) {
    if (field === null || field === undefined) {
        return '';
    }
    
    // Convert to string
    field = String(field);
    
    // If the field contains commas, quotes, or newlines, wrap it in quotes
    if (field.includes(',') || field.includes('"') || field.includes('\n')) {
        // Double any existing quotes
        field = field.replace(/"/g, '""');
        // Wrap in quotes
        field = '"' + field + '"';
    }
    
    return field;
}

var toggleLock = function (id) {
    api.groupId.lock(id)
        .success(function (group) {
            var locked = group.locked
            var btn = $("button[onclick='toggleLock(" + id + ")']")
            btn.removeClass('btn-default btn-warning')
               .addClass(locked ? 'btn-warning' : 'btn-default')
               .attr('title', locked ? 'Unlock group' : 'Lock group')
               .find('i')
               .removeClass('fa-lock fa-unlock-alt')
               .addClass(locked ? 'fa-lock' : 'fa-unlock-alt')
            var row = btn.closest('tr')
            var nameCell = row.find('td:first')
            var plainName = nameCell.text().trim()
            if (locked) {
                nameCell.html("<i class='fa fa-lock' style='margin-right:5px;color:#e6a817;' title='Locked — not available in campaign pickers'></i>" + escapeHtml(plainName))
            } else {
                nameCell.text(plainName)
            }
            // keep local groups array in sync
            var g = groups.find(function(x) { return x.id === id })
            if (g) g.locked = locked
        })
        .error(function () {
            errorFlash("Error updating group lock status")
        })
}

var deleteGroup = function (id) {
    var group = groups.find(function (x) {
        return x.id === id
    })
    if (!group) {
        return
    }
    Swal.fire({
        title: "Are you sure?",
        text: "This will delete the group. This can't be undone!",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete " + escapeHtml(group.name),
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function () {
            return new Promise(function (resolve, reject) {
                api.groupId.delete(id)
                    .success(function (msg) {
                        resolve()
                    })
                    .error(function (data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function (result) {
        if (result.value){
            Swal.fire(
                'Group Deleted!',
                'This group has been deleted!',
                'success'
            );
        }
        $('button:contains("OK")').on('click', function () {
            location.reload()
        })
    })
}

function addTarget(firstNameInput, lastNameInput, emailInput, phoneInput, positionInput, customInput,
    departmentInput, companyInput, cityInput, stateInput, countryInput, unitInput, tagsInput) {
    // Create new data row.
    var email = emailInput ? escapeHtml(emailInput).toLowerCase() : "";
    var phone = phoneInput ? escapeHtml(phoneInput) : "";
    var newRow = [
        escapeHtml(firstNameInput),
        escapeHtml(lastNameInput),
        email,
        phone,
        escapeHtml(positionInput),
        escapeHtml(customInput),
        escapeHtml(departmentInput || ""),
        escapeHtml(companyInput || ""),
        escapeHtml(cityInput || ""),
        escapeHtml(stateInput || ""),
        escapeHtml(countryInput || ""),
        escapeHtml(unitInput || ""),
        escapeHtml(tagsInput || ""),
        '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
    ];

    // Check table to see if email or phone already exists.
    var targetsTable = targets.DataTable();
    var existingRowIndex = -1;
    
    // First check if we have a match by both email and phone (if both are provided)
    if (email && phone) {
        // console.log("Checking for target with both email and phone:", email, phone);
        targetsTable.rows().every(function(rowIdx) {
            var rowData = this.data();
            if (rowData[2] === email && rowData[3] === phone) {
                // console.log("Found match by both email and phone at row:", rowIdx);
                existingRowIndex = rowIdx;
                return false; // Break the loop
            }
            return true;
        });
    }
    
    // If no match found by both, check if email exists
    if (existingRowIndex < 0 && email) {
        // console.log("Checking for target with email:", email);
        targetsTable.rows().every(function(rowIdx) {
            var rowData = this.data();
            if (rowData[2] === email) {
                // console.log("Found match by email at row:", rowIdx);
                existingRowIndex = rowIdx;
                return false; // Break the loop
            }
            return true;
        });
    }
    
    // If still no match found and phone is provided, check if phone exists
    if (existingRowIndex < 0 && phone) {
        // console.log("Checking for target with phone:", phone);
        targetsTable.rows().every(function(rowIdx) {
            var rowData = this.data();
            if (rowData[3] === phone) {
                // console.log("Found match by phone at row:", rowIdx);
                existingRowIndex = rowIdx;
                return false; // Break the loop
            }
            return true;
        });
    }
    
    // Update or add new row as necessary.
    if (existingRowIndex >= 0) {
        targetsTable
            .row(existingRowIndex, {
                order: "index"
            })
            .data(newRow);
    } else {
        targetsTable.row.add(newRow);
    }
}

function load() {
    $("#groupTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.groups.summary()
        .success(function (response) {
            // console.log("Group summary response:", response);
            $("#loading").hide()
            if (response.total > 0) {
                groups = response.groups
                // console.log("Groups data:", groups);
                $("#emptyMessage").hide()
                $("#groupTable").show()
                var groupTable = $("#groupTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                groupTable.clear();
                groupRows = []
                $.each(groups, function (i, group) {
                    var lockIcon = group.locked ? 'fa-lock' : 'fa-unlock-alt'
                    var lockTitle = group.locked ? 'Unlock group' : 'Lock group'
                    var lockBtnClass = group.locked ? 'btn-warning' : 'btn-default'
                    var nameCell = group.locked
                        ? "<i class='fa fa-lock' style='margin-right:5px;color:#e6a817;' title='Locked — not available in campaign pickers'></i>" + escapeHtml(group.name)
                        : escapeHtml(group.name)
                    groupRows.push([
                        nameCell,
                        escapeHtml(group.num_targets),
                        moment(group.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='pull-right'>\
                    <button class='btn btn-primary' data-toggle='modal' data-backdrop='static' data-target='#modal' onclick='edit(" + group.id + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-primary' onclick='downloadGroup(" + group.id + ")'>\
                    <i class='fa fa-download'></i>\
                    </button>\
                    <button class='btn " + lockBtnClass + "' onclick='toggleLock(" + group.id + ")' title='" + lockTitle + "'>\
                    <i class='fa " + lockIcon + "'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='deleteGroup(" + group.id + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ])
                })
                groupTable.rows.add(groupRows).draw()
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function () {
            errorFlash("Error fetching groups")
        })
}

$(document).ready(function () {
    load()
    // Setup the event listeners
    // Handle manual additions
    $("#targetForm").submit(function () {
        // Validate the form data
        var targetForm = document.getElementById("targetForm")
        if (!targetForm.checkValidity()) {
            targetForm.reportValidity()
            return
        }
        
        // Check if either email or phone is provided
        var email = $("#email").val();
        var phone = $("#phone").val();
        if (!email && !phone) {
            modalError("Either Email or Phone must be provided");
            return false;
        }
        
        addTarget(
            $("#firstName").val(),
            $("#lastName").val(),
            email,
            phone,
            $("#position").val(),
            $("#custom").val(),
            $("#department").val(),
            $("#company").val(),
            $("#city").val(),
            $("#state").val(),
            $("#country").val(),
            $("#unit").val(),
            $("#tags").val());
        targets.DataTable().draw();
        refreshFilterOptions();
        validateTargetRows();

        // Reset user input.
        $("#targetForm>div>input").val('');
        $("#firstName").focus();
        return false;
    });
    // Handle Deletion
    $("#targetsTable").on("click", "span>i.fa-trash-o", function () {
        targets.DataTable()
            .row($(this).parents('tr'))
            .remove()
            .draw();
        refreshFilterOptions();
        validateTargetRows();
    });
    $("#modal").on("hide.bs.modal", function () {
        dismiss();
    });
    $("#csv-template").click(downloadCSVTemplate)

    $("#filterCompany, #filterDepartment, #filterCity, #filterState, #filterCountry").on('change', function () {
        applyTargetFilters();
    });
    $("#filterTag").on('keyup', function () {
        applyTargetFilters();
    });
});

// Column indices in the #targetsTable DataTable, per the order built by
// addTarget()/edit(): first_name, last_name, email, phone, position,
// custom, department, company, city, state, country, unit, tags, actions.
var TARGET_COLUMN = {
    EMAIL: 2,
    PHONE: 3,
    DEPARTMENT: 6,
    COMPANY: 7,
    CITY: 8,
    STATE: 9,
    COUNTRY: 10,
    UNIT: 11,
    TAGS: 12
};

var emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
var phonePattern = /^\+?[0-9]{7,15}$/;

// Flags rows with an invalid email/phone or a duplicate email, and updates
// the import summary line above the grid.
function validateTargetRows() {
    var seenEmails = {};
    var invalidCount = 0;
    var duplicateCount = 0;
    targets.DataTable().rows().every(function () {
        var data = this.data();
        var node = $(this.node());
        node.removeClass('row-invalid row-duplicate');
        var email = unescapeHtml(data[TARGET_COLUMN.EMAIL] || "");
        var phone = unescapeHtml(data[TARGET_COLUMN.PHONE] || "");
        var isValid = emailPattern.test(email) && (phone === '' || phonePattern.test(phone));
        if (!isValid) {
            node.addClass('row-invalid');
            invalidCount++;
        }
        var emailKey = email.toLowerCase();
        if (emailKey) {
            if (seenEmails[emailKey]) {
                node.addClass('row-duplicate');
                duplicateCount++;
            }
            seenEmails[emailKey] = true;
        }
    });
    var total = targets.DataTable().rows().count();
    var needsAttention = invalidCount + duplicateCount;
    if (needsAttention > 0) {
        $("#importSummary").text(total + " imported, " + needsAttention + " need attention");
    } else if (total > 0) {
        $("#importSummary").text(total + " imported");
    } else {
        $("#importSummary").text("");
    }
}

// Rebuilds the segmentation filter dropdown options from the values
// currently present in the grid.
function refreshFilterOptions() {
    var companies = {}, departments = {}, cities = {}, states = {}, countries = {};
    targets.DataTable().rows().every(function () {
        var data = this.data();
        var department = unescapeHtml(data[TARGET_COLUMN.DEPARTMENT] || "");
        var company = unescapeHtml(data[TARGET_COLUMN.COMPANY] || "");
        var city = unescapeHtml(data[TARGET_COLUMN.CITY] || "");
        var state = unescapeHtml(data[TARGET_COLUMN.STATE] || "");
        var country = unescapeHtml(data[TARGET_COLUMN.COUNTRY] || "");
        if (department) departments[department] = true;
        if (company) companies[company] = true;
        if (city) cities[city] = true;
        if (state) states[state] = true;
        if (country) countries[country] = true;
    });
    function fillSelect(selector, values, placeholder) {
        var select = $(selector);
        var current = select.val();
        select.empty().append($('<option value="">' + placeholder + '</option>'));
        Object.keys(values).sort().forEach(function (v) {
            select.append($('<option></option>').attr('value', v).text(v));
        });
        select.val(current || "");
    }
    fillSelect('#filterCompany', companies, 'All Companies');
    fillSelect('#filterDepartment', departments, 'All Departments');
    fillSelect('#filterCity', cities, 'All Cities');
    fillSelect('#filterState', states, 'All States');
    fillSelect('#filterCountry', countries, 'All Countries');
}

// Applies the segmentation filter controls to the targets DataTable.
function applyTargetFilters() {
    var dt = targets.DataTable();
    var exact = function (value) {
        return value ? '^' + $.fn.dataTable.util.escapeRegex(value) + '$' : '';
    };
    dt.column(TARGET_COLUMN.COMPANY).search(exact($("#filterCompany").val()), true, false);
    dt.column(TARGET_COLUMN.DEPARTMENT).search(exact($("#filterDepartment").val()), true, false);
    dt.column(TARGET_COLUMN.CITY).search(exact($("#filterCity").val()), true, false);
    dt.column(TARGET_COLUMN.STATE).search(exact($("#filterState").val()), true, false);
    dt.column(TARGET_COLUMN.COUNTRY).search(exact($("#filterCountry").val()), true, false);
    dt.column(TARGET_COLUMN.TAGS).search($("#filterTag").val() || '');
    dt.draw();
}

var trainings = []
var currentTrainingId = null

function newTraining() {
    currentTrainingId = null
    $("#trainingModalLabel").text("New Training")
    $("#trainingId").val("")
    $("#trainingName").val("")
    $("#trainingDescription").val("")
    $("#lessonTable tbody").empty()
    $("#questionTable tbody").empty()
    $("#quizPassPercent").val(80)
    $("#quizMaxAttempts").val(0)
    loadGroupsForAssign()
}

function edit(id) {
    if (id === -1) {
        newTraining()
        return
    }
    currentTrainingId = id
    var t = trainings.find(function (t) { return t.id === id })
    $("#trainingModalLabel").text("Edit Training: " + t.name)
    $("#trainingId").val(t.id)
    $("#trainingName").val(t.name)
    $("#trainingDescription").val(t.description || "")
    renderLessons(t.lessons || [])
    renderQuestions(t.quiz ? (t.quiz.questions || []) : [])
    $("#quizPassPercent").val(t.quiz ? t.quiz.pass_percent : 80)
    $("#quizMaxAttempts").val(t.quiz ? t.quiz.max_attempts : 0)
    loadGroupsForAssign()
}

function renderLessons(lessons) {
    var tbody = $("#lessonTable tbody")
    tbody.empty()
    lessons.forEach(function (l) {
        tbody.append("<tr><td>" + l.position + "</td><td>" + escapeHtml(l.title) +
            "</td><td><button type=\"button\" class=\"btn btn-danger btn-xs\" onclick=\"removeLesson(" + l.id + ")\"><i class=\"fa fa-trash-o\"></i></button></td></tr>")
    })
}

function renderQuestions(questions) {
    var tbody = $("#questionTable tbody")
    tbody.empty()
    questions.forEach(function (q) {
        tbody.append("<tr><td>" + q.position + "</td><td>" + escapeHtml(q.type) + "</td><td>" + escapeHtml(q.text) + "</td><td></td></tr>")
    })
}

function addLesson() {
    var title = $("#newLessonTitle").val()
    var html = $("#newLessonHTML").val()
    if (!title) {
        errorFlash("Lesson title is required")
        return
    }
    if (!currentTrainingId) {
        errorFlash("Save the training before adding lessons")
        return
    }
    api.trainingId.addLesson(currentTrainingId, { title: title, html: html }).success(function () {
        $("#newLessonTitle").val("")
        $("#newLessonHTML").val("")
        edit(currentTrainingId)
        load()
    }).error(function () {
        errorFlash("Error adding lesson")
    })
}

function removeLesson(lessonId) {
    api.trainingId.deleteLesson(currentTrainingId, lessonId).success(function () {
        edit(currentTrainingId)
        load()
    })
}

function saveQuizSettings() {
    if (!currentTrainingId) {
        errorFlash("Save the training before configuring the quiz")
        return
    }
    api.trainingId.saveQuiz(currentTrainingId, {
        pass_percent: parseInt($("#quizPassPercent").val(), 10),
        max_attempts: parseInt($("#quizMaxAttempts").val(), 10)
    }).success(function () {
        successFlash("Quiz settings saved")
        edit(currentTrainingId)
        load()
    }).error(function () {
        errorFlash("Error saving quiz settings")
    })
}

function addQuestion() {
    if (!currentTrainingId) {
        errorFlash("Save the training before adding quiz questions")
        return
    }
    var type = $("#newQuestionType").val()
    var text = $("#newQuestionText").val()
    var options = $("#newQuestionOptions").val().split(",").map(function (s) { return s.trim() }).filter(Boolean)
    var answer = $("#newQuestionAnswer").val()
    if (!text || !answer) {
        errorFlash("Question text and correct answer are required")
        return
    }
    api.trainingId.addQuestion(currentTrainingId, {
        type: type,
        text: text,
        options: JSON.stringify(type === "multiple_choice" ? options : []),
        correct_answer: answer
    }).success(function () {
        $("#newQuestionText").val("")
        $("#newQuestionOptions").val("")
        $("#newQuestionAnswer").val("")
        edit(currentTrainingId)
        load()
    }).error(function () {
        errorFlash("Error adding question")
    })
}

function loadGroupsForAssign() {
    api.groups.get().success(function (gs) {
        var select = $("#assignGroupSelect")
        select.empty()
        gs.forEach(function (g) {
            select.append("<option value=\"" + g.id + "\">" + escapeHtml(g.name) + "</option>")
        })
    })
}

function assignToGroup() {
    if (!currentTrainingId) {
        errorFlash("Save the training before assigning it")
        return
    }
    var groupId = parseInt($("#assignGroupSelect").val(), 10)
    if (!groupId) {
        errorFlash("Choose a group to assign to")
        return
    }
    api.trainingId.assign(currentTrainingId, groupId).success(function (data) {
        successFlash(data.message || "Training assigned")
    }).error(function () {
        errorFlash("Error assigning training")
    })
}

function save() {
    var training = {
        name: $("#trainingName").val(),
        description: $("#trainingDescription").val()
    }
    var id = $("#trainingId").val()
    var request
    if (id) {
        training.id = parseInt(id)
        request = api.trainingId.put(training)
    } else {
        request = api.trainings.post(training)
    }
    request.success(function (data) {
        successFlash("Training saved successfully!")
        currentTrainingId = data.id
        $("#trainingId").val(data.id)
        load()
        edit(currentTrainingId)
    }).error(function () {
        errorFlash("Error saving training")
    })
}

function deleteTraining(id) {
    if (!confirm("Delete this training? This also removes its lessons and quiz.")) {
        return
    }
    api.trainingId.delete(id).success(function () {
        successFlash("Training deleted")
        load()
    })
}

function load() {
    $("#loading").show()
    $("#trainingTable").hide()
    api.trainings.get().success(function (ts) {
        $("#loading").hide()
        trainings = ts
        if (ts.length === 0) {
            $("#emptyMessage").show()
            return
        }
        $("#emptyMessage").hide()
        $("#trainingTable").show()
        if ($.fn.DataTable.isDataTable("#trainingTable")) {
            $("#trainingTable").DataTable().destroy()
        }
        var rows = ts.map(function (t) {
            return [
                escapeHtml(t.name),
                (t.lessons || []).length,
                t.quiz ? ((t.quiz.questions || []).length + " questions") : "None",
                "<div class=\"pull-right\">" +
                "<button type=\"button\" class=\"btn btn-primary\" data-toggle=\"modal\" data-target=\"#trainingModal\" onclick=\"edit(" + t.id + ")\"><i class=\"fa fa-pencil\"></i></button>&nbsp;" +
                "<button type=\"button\" class=\"btn btn-danger\" onclick=\"deleteTraining(" + t.id + ")\"><i class=\"fa fa-trash-o\"></i></button>" +
                "</div>"
            ]
        })
        $("#trainingTable").DataTable({
            data: rows,
            destroy: true,
            columnDefs: [{ orderable: false, targets: "no-sort" }]
        })
    }).error(function () {
        $("#loading").hide()
        errorFlash("Error fetching trainings")
    })
}

$(document).ready(function () {
    load()
})

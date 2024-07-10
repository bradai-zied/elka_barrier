$(document).ready(function() {
  var table = $('#barriers-table').DataTable({
      responsive: true,
      ajax: {
          url: '/Barrier',
          dataSrc: 'barriers'
      },
      columns: [
          {data: 'name', className: 'editable'},
          {data: 'ip', className: 'editable'},
          {data: 'id', className: 'editable'},
          {data: 'barrierType', className: 'editable'},
          {data: 'port', className: 'editable'},
          {
              data: null,
              defaultContent: '',
              orderable: false,
              className: 'actions'
          }
      ],
      dom: '<"top"Bf>rt<"bottom"lip>',
      select: true,
      buttons: [
          {
              text: 'Add',
              action: function () {
                  var newRow = {
                      name: '',
                      ip: '',
                      id: '',
                      barrierType: '',
                      port: '52719'  // Default port value
                  };
                  var addedRow = table.row.add(newRow).draw().node();
                  $(addedRow).addClass('new-row');
                  addButtonsToRow(addedRow);
                  table.row(addedRow).select();
                  table.cell(addedRow, 0).focus();
              }
          }
      ],
      rowCallback: function(row, data, index) {
          if (data.id === '') {
              $(row).addClass('new-row');
          }
          addButtonsToRow(row);
      },
      drawCallback: function(settings) {
          $('.new-row').css('height', '60px');  // Increase height of new rows
      },
      columnDefs: [
          { className: "dt-center", targets: "_all" }
      ],
      language: {
          search: "_INPUT_",
          searchPlaceholder: "Search barriers..."
      }
  });

  function addButtonsToRow(row) {
      var actionsCell = $('td.actions', row);
      if (actionsCell.children().length === 0) {
          actionsCell.html('<button class="btn btn-primary btn-sm save">Save</button> <button class="btn btn-danger btn-sm delete">Delete</button>');
      }
  }

  // Handle inline editing
  $('#barriers-table').on('click', 'td.editable', function(e) {
      e.stopPropagation();
      var cell = table.cell(this);
      var originalValue = cell.data();
      var input = $('<input type="text" class="form-control">').val(originalValue);
      $(this).html(input);
      input.focus();
  });

  $('#barriers-table').on('blur', 'td.editable input', function() {
      var cell = table.cell($(this).parent());
      var newValue = $(this).val();
      cell.data(newValue).draw();
  });

  // Handle barrierType as a dropdown
  $('#barriers-table').on('click', 'td.editable:nth-child(4)', function(e) {
      e.stopPropagation();
      var cell = table.cell(this);
      var currentValue = cell.data();
      var select = $('<select class="form-control"><option value="entry">entry</option><option value="exit">exit</option></select>');
      select.val(currentValue);
      $(this).html(select);
      select.focus();
  });

  $('#barriers-table').on('change', 'td.editable:nth-child(4) select', function() {
      var cell = table.cell($(this).parent());
      var newValue = $(this).val();
      cell.data(newValue).draw();
  });

  // Handle Save button click
  $('#barriers-table').on('click', '.save', function() {
      var row = table.row($(this).parents('tr'));
      var rowData = row.data();
      rowData.port = parseInt(rowData.port, 10);
      if (rowData.id !== '') {
          rowData.id = parseInt(rowData.id, 10);
      }
      if ($(row.node()).hasClass('new-row')) {
          // This is a new row, send POST request
          $.ajax({
              url: '/add',
              method: 'POST',
              contentType: 'application/json',
              data: JSON.stringify(rowData),
              success: function(response) {
                  toastr.success('New barrier added successfully');
                  $(row.node()).removeClass('new-row');
                  row.data(response).draw();
              },
              error: function(xhr, status, error) {
                  toastr.error('Error adding new barrier');
              }
          });
      } else {
          // This is an existing row, send PUT request to update
          $.ajax({
              url: '/update/' + rowData.id,
              method: 'PUT',
              contentType: 'application/json',
              data: JSON.stringify(rowData),
              success: function(response) {
                  toastr.success('Barrier updated successfully');
                  row.data(response).draw();
              },
              error: function(xhr, status, error) {
                  toastr.error('Error updating barrier');
              }
          });
      }
  });

  // Handle Delete button click
  $('#barriers-table').on('click', '.delete', function() {
      var row = table.row($(this).parents('tr'));
      var rowData = row.data();
      
      if (confirm('Are you sure you want to delete this barrier?')) {
          if ($(row.node()).hasClass('new-row')) {
              // This is a new, unsaved row. Just remove it.
              row.remove().draw();
              toastr.success('New barrier removed');
          } else {
              // This is an existing row, send DELETE request
              $.ajax({
                  url: '/delete/' + rowData.id,
                  method: 'DELETE',
                  success: function(response) {
                      toastr.success('Barrier deleted successfully');
                      row.remove().draw();
                  },
                  error: function(xhr, status, error) {
                      toastr.error('Error deleting barrier');
                  }
              });
          }
      }
  });



  // Restart Application button functionality
  $('#restart-app').on('click', function() {
    if (confirm('Are you sure you want to restart the application?')) {
        $.ajax({
            url: '/restart',
            method: 'GET',
            success: function(response) {
                toastr.success('Application restart initiated.');
                // Optionally, you can redirect to a specific page or reload the current page after a delay
                setTimeout(function() {
                    window.location.reload();
                }, 5000);  // Reload after 5 seconds
            },
            error: function(xhr, status, error) {
                toastr.error('Failed to restart the application.');
            }
        });
    }
});
});
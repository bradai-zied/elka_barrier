$(document).ready(function() {
  fetchBarrierStatus();

  // Refresh data every 30 seconds
  setInterval(fetchBarrierStatus, 30000);

  function fetchBarrierStatus() {
      $.ajax({
          url: '/allstatus',
          method: 'GET',
          dataType: 'json',
          success: function(data) {
              updateDashboard(data);
          },
          error: function(xhr, status, error) {
              console.error("Error fetching barrier status:", error);
              toastr.error("Failed to fetch barrier status");
          }
      });
  }

  function updateDashboard(data) {
      $('.barrier-grid').empty();  // Clear existing barriers

      for (let controllerType in data) {
          for (let barrierIp in data[controllerType]) {
              let barrierData = data[controllerType][barrierIp];
              let barrierHtml = createBarrierCard(barrierData);
              $('.barrier-grid').append(barrierHtml);
          }
      }

      // Reattach event listeners for buttons
      attachButtonListeners();
  }

  function createBarrierCard(barrierData) {
      let barrierImage = barrierData.IsConnected
          ? `<img src="/static/img/barrier-${barrierData.IsClosed ? 'closed' : 'open'}.svg" alt="Barrier ${barrierData.BarrierPositionStr}">`
          : '<p>No image available</p>';

      let buttonClass = barrierData.IsConnected ? '' : 'disabled';

      return `
          <div class="barrier-card" data-id="${barrierData.Id}">
              <div class="barrier-status">
                  <div class="status-info">
                      <h2>Barrier ${barrierData.Id}</h2>
                      <p class="status ${barrierData.IsConnected ? 'connected' : 'disconnected'}">
                          ${barrierData.IsConnected ? 'Connected' : 'Disconnected'}
                      </p>
                      <p class="lock-status">Locked Up: <span class="${barrierData.IsLockedUp ? 'true' : 'false'}">
                          ${barrierData.IsLockedUp ? 'Yes' : 'No'}
                      </span></p>
                      <p class="lock-status">Locked Down: <span class="${barrierData.IsLockedDown ? 'true' : 'false'}">
                          ${barrierData.IsLockedDown ? 'Yes' : 'No'}
                      </span></p>
                      <p class="ip-address">IP: ${barrierData.Barrierip}</p>
                  </div>
                  <div class="barrier-image">
                      ${barrierImage}
                  </div>
              </div>
              <div class="barrier-controls">
                  <button class="btn btn-open ${buttonClass}">Open</button>
                  <button class="btn btn-close ${buttonClass}">Close</button>
                  <button class="btn btn-unlock ${buttonClass}">Unlock</button>
                  <button class="btn btn-lockup ${buttonClass}">Lock Up</button>
                  <button class="btn btn-lockdown ${buttonClass}">Lock Down</button>
                  <button class="btn btn-sync ${buttonClass}">Sync</button>
              </div>
          </div>
      `;
  }

  function attachButtonListeners() {
      $('.barrier-controls .btn:not(.disabled)').off('click').on('click', function() {
          let action = $(this).text().toLowerCase();
          let barrierId = $(this).closest('.barrier-card').data('id');
          sendBarrierCommand(barrierId, action);
      });
  }

  function sendBarrierCommand(barrierId, action) {
      let url, method;
      switch(action) {
          case 'open':
              url = `/open/${barrierId}`;
              method = 'POST';
              break;
          case 'close':
              url = `/close/${barrierId}`;
              method = 'POST';
              break;
          case 'lock up':
              url = `/open/${barrierId}?lock=true`;
              method = 'POST';
              break;
          case 'lock down':
              url = `/close/${barrierId}?lock=true`;
              method = 'POST';
              break;
          case 'unlock':
              url = `/unlock/${barrierId}`;
              method = 'POST';
              break;
          case 'sync':
              url = `/query/${barrierId}?query=2`;
              method = 'GET';
              break;
          default:
              console.error('Unknown action:', action);
              toastr.error('Unknown action');
              return;
      }

      $.ajax({
          url: url,
          method: method,
          success: function(response) {
              toastr.success(`Command ${action} sent successfully to Barrier ${barrierId}`);
              fetchBarrierStatus();  // Refresh data after command
          },
          error: function(xhr, status, error) {
              toastr.error(`Failed to send command ${action} to Barrier ${barrierId}`);
          }
      });
  }
  // Add this inside your $(document).ready(function() { ... });

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
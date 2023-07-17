(function () {
  const el = document.querySelector('.menu');
})();

// Attach an event listener to the window
window.addEventListener('load', function() {
  // Get the form element
  var form = document.getElementById('downloadform');
    // Attach an event listener to the form
    form.addEventListener('submit', function(event) {
        // Prevent the default form submission behavior
        event.preventDefault();
        
        // Get the value of the input field
        var inputValue = form.elements['link-input'].value;
        
        // Do something with the value...
        sendLink(inputValue);
    });
});


function addActiveClass(context) {
  // menu icons 
  
  navItem = document.getElementById(context);
  navItems.forEach((nav) => {
    nav.classList.remove('nav-active');
  });
  navItem.classList.add('nav-active');

  //view toggle
/*
- click menu option
- find all views and remove active class(which sets opacity)
- add active class to selected view
- run function for pulling data and displaying in view

*/
  const allviews = document.querySelectorAll("[class*=' view']");
  const selectedview = document.querySelector('[class*=' + context + ']');
  // console.log('[class$=' + context + ']');

  allviews.forEach((view) => {
    // console.log(view);
    view.classList.remove('viewactive');
  })
  // console.log(selectedview)
  selectedview.classList.add('viewactive')

}

function changeMenu(context) {
  const title = document.querySelector('.title');

  switch (context) {
    case 'home':
      title.innerText = 'Home';
      addActiveClass(context);
      break;
    case 'tasks':
      title.innerText = 'Tasks';
      addActiveClass(context);
      break;
    case 'download':
      title.innerText = 'Download';
      addActiveClass(context);
      break;
    case 'settings':
      title.innerText = 'Settings';
      addActiveClass(context);
      break;
    default:
      break;
  }
}

function addToolTip(key) {
  removeTooltips();
  const toolTips = document.querySelectorAll('.tooltip');
 
  toolTips.forEach((tooltip) => {
    if (tooltip.getAttribute('data-key') == key) {
      tooltip.style.opacity = '1';
    }
  });
}

function removeTooltips() {
  const toolTips = document.querySelectorAll('.tooltip');
  toolTips.forEach((tooltip) => {
    tooltip.style.opacity = '0';
  });
}

const navItems = document.querySelectorAll('.menu-icon');
navItems.forEach((item) => {
  item.addEventListener('mouseover', () => {
    addToolTip(item.getAttribute('id'));
  });
});

navItems.forEach((item) => {
  item.addEventListener('mouseleave', () => {
    removeTooltips();
  });
});

function SendData() {
  // var data = new FormData(document.getElementById("taskform"))
  // // console.log(data);
  // let post = JSON.stringify(data)
  
  // console.log(post);
  const url = "api/tasks"
  let xhr = new XMLHttpRequest()
   
  xhr.open('POST', url, true)
  xhr.setRequestHeader('Content-type', 'application/json; charset=UTF-8')
  xhr.send("Yes");
   
  xhr.onload = function () {
      if(xhr.status === 201) {
          console.log("Post successfully created!") 
      }
  }
}

function UpdateSettings() {
  var formData = new FormData(document.getElementById("settingsform"))
  // console.log(formData);
  let post = JSON.stringify(Object.fromEntries(formData));//JSON.stringify(data)
  
  // console.log(post);
  const url = "api/settings"
  let xhr = new XMLHttpRequest()
   
  xhr.open('POST', url, true)
  xhr.setRequestHeader('Content-type', 'application/json; charset=UTF-8')
  xhr.send(post);
   
  xhr.onload = function () {
      if(xhr.status === 201) {
          console.log("Post successfully created!") 
      }
  }
}

// Youtube Background Download

function sendLink_old(url) {
  // Get a reference to the submit button
  const submitButton = document.getElementById('submit-download');
  // Create a new XHR object
  const xhr = new XMLHttpRequest();
  // Open a new request
  xhr.open('post', 'api/download', true);

  // Set the request headers
  xhr.setRequestHeader('Content-Type', 'application/json');

  var data = {
    'link': url
  }

  // Handle the response
  xhr.onload = () => {
    if (xhr.status === 200) {
      document.getElementById('link-input').value = ''; 
      // Request was successful, handle the response
      const response = JSON.parse(xhr.response);
      // TODO: Use the response data
      console.log(response);
      // Enable the submit button
      submitButton.disabled = false;
    } else {
      // Request failed, handle the error
      console.error(`Request failed: ${xhr.status}`);
      // Enable the submit button
      submitButton.disabled = false;
    }
  };

  xhr.ontimeout = (e) => {
  // XMLHttpRequest timed out. Do something here.
    console.log("Error ontimeout");
};


  // Disable the submit button
  submitButton.disabled = true;

  // Send the request
  xhr.send(JSON.stringify(data));
}

function sendLink(url) {
  // Get a reference to the submit button
  const submitButton = document.getElementById('submit-download');
  // Create the request payload
  const data = {
    link: url
  };

  // Disable the submit button
  submitButton.disabled = true;

  // Send the request using Fetch API
  fetch('api/downloadyt', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  })
    .then(response => {
      if (response.ok) {
        // Request was successful, handle the response
        return response.json();
      } else {
        // Request failed, handle the error
        throw new Error(`Request failed: ${response.status}`);
      }
    })
    .then(responseData => {
      // TODO: Use the response data
      console.log(responseData);
      // Clear the input field
      document.getElementById('link-input').value = '';
    })
    .catch(error => {
      // Handle any errors that occurred during the request
      console.error(error);
    })
    .finally(() => {
      // Enable the submit button
      submitButton.disabled = false;
    });
}

document.addEventListener('submit', function(event) {
  event.preventDefault();
  const form = event.target;
  const endpoint = form.getAttribute('data-endpoint');
  
  const formData = new FormData(form);
  const requestData = {};
  for (let [name, value] of formData) {
    requestData[name] = value;
  }
  
  fetch(endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(requestData)
  })
    .then(response => {
      if (response.ok) {
        return response.json();
      } else {
        throw new Error(`Request failed: ${response.status}`);
      }
    })
    .then(responseData => {
      console.log(responseData);
      form.reset();
      // We want to show a modal with a list of files to select then send that response back to the server.
      // once the server selects the files and downloads, a notifcation will be sent.
      // We then may want to be able to download that directly and auto unrestrict that link instead of going to server?
      if('RD' in responseData['TorrentInfo'][files]){
        console.log('test')

        const filemodal = new bootstrap.Modal('#rdFileSelectionModal', {
          keyboard: true, backdrop: false, focus:true
        }).show()
        console.log(filemodal)
      }
    })
    .catch(error => {
      console.error(error);
    });
});


// Settings folder browser

function GetFolders(type) {

  const data = document.getElementById(type+"Field").value;
  fetch('api/getfolders?' + data)
    .then(res => res.json())
    .then(res => console.log(res))
    .catch(err => console.error(err));
}

// Initialize notification queue
const notificationQueue = [];

// Function to display a toast notification
function showToast(message) {
  const toast = document.createElement('div');
  toast.classList.add('notification');
  toast.textContent = message;

  // Create progress bar
  const progressBar = document.createElement('div');
  progressBar.classList.add('progress-bar','bg-info');
  progressBar.role = 'progressBar';
  toast.appendChild(progressBar);

  // Animate the progress bar
  const duration = 3000; // Adjust the duration as needed
  let startTime = Date.now();

  const animateProgressBar = () => {
    const currentTime = Date.now();
    const elapsedTime = currentTime - startTime;
    const progress = (elapsedTime / duration) * 100;

    progressBar.style.width = `${progress}%`;

    if (progress < 100) {
      requestAnimationFrame(animateProgressBar);
    } else {
      // Remove the toast when the progress is complete
      document.body.removeChild(toast);

      // Display the next notification in the queue
      if (notificationQueue.length > 0) {
        const nextNotification = notificationQueue.shift();
        showToast(nextNotification);
      }
    }
  };

  // Add the toast to the DOM
  document.body.appendChild(toast);

  // Start animating the progress bar
  requestAnimationFrame(animateProgressBar);
}
// Function to add a notification to the queue
function enqueueNotification(message) {
  notificationQueue.push(message);

  // If no notification is currently displayed, display the new notification
  if (notificationQueue.length === 1) {
    showToast(message);
  }
}

const eventSource = new EventSource("/api/handshake");

eventSource.addEventListener("close", function(event) {
  // Handle SSE close event
  console.log(event);
  alert("Connection with the server has been lost..");
}, false);

eventSource.onopen = (e) => {
  console.log("The connection has been established." , e);
  enqueueNotification(e.data)
  // displayNotification(e.data)
}
eventSource.onmessage = (e) => {
  console.log("New Message from the server!\n", e);
  enqueueNotification(JSON.parse(e.data)['message'])
  // displayNotification(JSON.parse(e.data)['message']);
}
eventSource.onerror = (e) => {
  // alert("Error! Connection with the server has been lost.. See logs for details");
  // console.log("Error: ", e);
  // console.log(eventSource)
  // if (eventSource.readyState == 2){
  //   enqueueNotification("Disconnected from server")
  //   // displayNotification("Disconnected from server")
  // }
  // if( == 1) {
  // displayNotification("Disconnected from server")
  // }
}
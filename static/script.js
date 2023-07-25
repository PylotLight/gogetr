(function () {
  const el = document.querySelector('.menu');
})();


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

// document.addEventListener('submit', function(event) {
//   event.preventDefault();
//   const form = event.target;
//   const endpoint = form.getAttribute('data-endpoint');
  
//   const formData = new FormData(form);
//   const requestData = {};
//   for (let [name, value] of formData) {
//     requestData[name] = value;
//   }

//   fetch(endpoint, {
//     method: 'POST',
//     headers: {
//       'Content-Type': 'application/json'
//     },
//     body: JSON.stringify(requestData)
//   })
//     .then(response => {
//       if (response.ok) {
//         return response.json();
//       } else {
//         throw new Error(`Request failed: ${response.status}`);
//       }
//     })
//     .then(responseData => {
//       console.log(responseData);
//       form.reset();
//       // We want to show a modal with a list of files to select then send that response back to the server.
//       // once the server selects the files and downloads, a notifcation will be sent.
//       // We then may want to be able to download that directly and auto unrestrict that link instead of going to server?
//       if('RD' in responseData['TorrentInfo'][files]){
//         console.log('test')

//         const filemodal = new bootstrap.Modal('#rdFileSelectionModal', {
//           keyboard: true, backdrop: false, focus:true
//         }).show()
//         console.log(filemodal)
//       }
//     })
//     .catch(error => {
//       console.error(error);
//     });
// });


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
  responseJson = JSON.parse(e.data)
  console.log("New Message from the server!\n", responseJson);
  if(responseJson['message'].includes("Next scan")){
    document.getElementById("ScanTime").innerText = responseJson['message'];
    return
  } 
  enqueueNotification(responseJson['message'])
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
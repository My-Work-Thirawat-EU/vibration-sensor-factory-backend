import tkinter as tk
from tkinter import ttk, messagebox
import requests

class LoginApp:
    def __init__(self, root):
        self.root = root
        self.root.title("Vibration Sensor Login")
        self.root.geometry("400x300")
        self.root.resizable(False, False)
        
        # Mockup user credentials
        self.mockup_users = {
            "admin": "admin123",
            "operator": "operator123",
            "viewer": "viewer123"
        }
        
        # Configure style
        self.style = ttk.Style()
        self.style.configure('TLabel', font=('Arial', 12))
        self.style.configure('TButton', font=('Arial', 12))
        self.style.configure('TEntry', font=('Arial', 12))
        
        # Create main frame
        self.main_frame = ttk.Frame(root, padding="20")
        self.main_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))
        
        # Title
        self.title_label = ttk.Label(
            self.main_frame, 
            text="Vibration Sensor System", 
            font=('Arial', 16, 'bold')
        )
        self.title_label.grid(row=0, column=0, columnspan=2, pady=20)
        
        # Username
        self.username_label = ttk.Label(self.main_frame, text="Username:")
        self.username_label.grid(row=1, column=0, sticky=tk.W, pady=5)
        self.username_entry = ttk.Entry(self.main_frame, width=30)
        self.username_entry.grid(row=1, column=1, pady=5)
        
        # Password
        self.password_label = ttk.Label(self.main_frame, text="Password:")
        self.password_label.grid(row=2, column=0, sticky=tk.W, pady=5)
        self.password_entry = ttk.Entry(self.main_frame, width=30, show="*")
        self.password_entry.grid(row=2, column=1, pady=5)
        
        # Login button
        self.login_button = ttk.Button(
            self.main_frame, 
            text="Login", 
            command=self.login
        )
        self.login_button.grid(row=3, column=0, columnspan=2, pady=20)
        
        # Mockup credentials info
        self.credential_info = ttk.Label(
            self.main_frame,
            text="Mockup Users:\nadmin/admin123\noperator/operator123\nviewer/viewer123",
            font=('Arial', 10),
            justify=tk.LEFT
        )
        self.credential_info.grid(row=4, column=0, columnspan=2, pady=10, sticky=tk.W)
        
        # Center the window
        self.root.update_idletasks()
        width = self.root.winfo_width()
        height = self.root.winfo_height()
        x = (self.root.winfo_screenwidth() // 2) - (width // 2)
        y = (self.root.winfo_screenheight() // 2) - (height // 2)
        self.root.geometry(f'{width}x{height}+{x}+{y}')
    
    def login(self):
        username = self.username_entry.get()
        password = self.password_entry.get()
        
        if not username or not password:
            messagebox.showerror("Error", "Please enter both username and password")
            return
        
        # Check against mockup users
        if username in self.mockup_users and self.mockup_users[username] == password:
            messagebox.showinfo("Success", f"Welcome, {username}!")
            # Here you would typically open the main application
        else:
            messagebox.showerror("Error", "Invalid username or password")

if __name__ == "__main__":
    root = tk.Tk()
    app = LoginApp(root)
    root.mainloop() 
import React from "react";
import "./LandingPage.css";

export default function LandingPage() {
  return (
    <>
      <section class="section-a">
        <div class="container">
          <div>
            <h2>The Okay File System </h2>
            <p>
              The Okay File System (OFS), inspired by the Google File System
              (GFS), is a distributed file system that makes it conveninent to
              share information and files among users. The scalable, highly
              available and eventually consistent properties of the OFS allow
              multiple users to perform any number of read and write operations
              concurrently.
            </p>
            <div>
              <h3>Features </h3>
              <ul style={{ paddingLeft: "50px" }}>
                <li>Atomic Append</li>
                <li>Read</li>
                <li>Chunking</li>
                <li>Replication</li>
                <li>Heartbeat</li>
              </ul>
            </div>
            <a
              href="https://github.com/KiranM27/OkayFileSystyem"
              class="btn"
              target="_blank"
            >
              OFS GitHub
            </a>{" "}
            &nbsp;
            <a
              href="https://static.googleusercontent.com/media/research.google.com/en//archive/gfs-sosp2003.pdf"
              class="btn"
              target="_blank"
            >
              GFS Paper
            </a>
          </div>
          <img
            src="https://blog.postman.com/wp-content/uploads/2021/03/Default-Blog-Image-1.png"
            alt=""
          />
        </div>
      </section>
    </>
  );
}
